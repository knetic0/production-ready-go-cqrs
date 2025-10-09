package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/contrib/otelfiber"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/knetic0/production-ready-go-cqrs/app/auth"
	"github.com/knetic0/production-ready-go-cqrs/app/healthcheck"
	"github.com/knetic0/production-ready-go-cqrs/app/user"
	"github.com/knetic0/production-ready-go-cqrs/infrastructure"
	"github.com/knetic0/production-ready-go-cqrs/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
)

// Prometheus metrics
var httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Duration of HTTP requests in seconds",
	Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
}, []string{"route", "method", "status"})

func init() {
	prometheus.MustRegister(httpRequestDuration)
}

func RequestDurationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Call next handler
		err := c.Next()

		// Record duration after the handler returns
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())

		// Record metrics
		httpRequestDuration.WithLabelValues(
			c.Route().Path,
			c.Method(),
			status,
		).Observe(duration)

		return err
	}
}

type Request any
type Response any

type HandlerInterface[TReq Request, TRes Response] interface {
	Handle(ctx context.Context, request *TReq) (*TRes, error)
}

var validate = validator.New()

func handle[TReq Request, TRes Response](handler HandlerInterface[TReq, TRes]) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req TReq

		if err := c.BodyParser(&req); err != nil && !errors.Is(err, fiber.ErrUnprocessableEntity) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.ParamsParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.QueryParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := c.ReqHeaderParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		if err := validate.Struct(req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		ctx := c.UserContext()
		res, err := handler.Handle(ctx, &req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(res)
	}
}

func main() {
	applicationConfig := config.Read()
	defer zap.L().Sync()
	zap.L().Info("app starting...")
	zap.L().Info("app config", zap.Any("appConfig", applicationConfig))

	tp := initTracer(applicationConfig)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tp.Shutdown(ctx); err != nil {
			zap.L().Error("failed to shutdown tracer provider", zap.Error(err))
		}
	}()
	client := httpc()

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = client.Transport
	retryClient.RetryMax = 0
	retryClient.RetryWaitMin = 100 * time.Millisecond
	retryClient.RetryWaitMax = 10 * time.Second
	retryClient.Backoff = retryablehttp.LinearJitterBackoff
	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Concurrency:  256 * 1024,
	})

	app.Use(otelfiber.Middleware())
	app.Use(RequestDurationMiddleware())

	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	db := infrastructure.NewPostgreAdapter(applicationConfig.Postgre.DSN)
	healthCheckHandler := healthcheck.NewHealthCheckHandler()
	userRepository := infrastructure.NewUserRepositoryAdapter(db)
	refreshTokenRepository := infrastructure.NewRefreshTokenRepositoryAdapter(db)
	userCreateHandler := user.NewUserCreateHandler(userRepository)
	userGetHandler := user.NewUserGetHandler(userRepository)
	userListHandler := user.NewUserListHandler(userRepository)
	loginHandler := auth.NewLoginHandler(userRepository, refreshTokenRepository, applicationConfig.Security)
	meHandler := user.NewMeHandler(userRepository)

	app.Post("/login/", handle(loginHandler))

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			JWTAlg: jwtware.HS256,
			Key:    []byte(applicationConfig.Security.JwtSecretKey),
		},
		SuccessHandler: func(c *fiber.Ctx) error {
			token := c.Locals("user").(*jwt.Token)
			claims := token.Claims.(jwt.MapClaims)
			if sub, err := claims.GetSubject(); err == nil {
				ctx := context.WithValue(c.UserContext(), "userId", sub)
				c.SetUserContext(ctx)
			}
			return c.Next()
		},
	}))

	app.Get("/healthcheck", handle(healthCheckHandler))
	app.Post("/users/", handle(userCreateHandler))
	app.Get("/users/:id", handle(userGetHandler))
	app.Get("/users/", handle(userListHandler))
	app.Get("/user", handle(meHandler))

	go func() {
		if err := app.Listen(fmt.Sprintf("0.0.0.0:%s", applicationConfig.Server.Port)); err != nil {
			zap.L().Error("Failed to start server", zap.Error(err))
			os.Exit(1)
		}
	}()

	zap.L().Info("Server started on port", zap.String("port", applicationConfig.Server.Port))
	gracefulShutdown(app)
}

func gracefulShutdown(app *fiber.App) {
	// Create channel for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	zap.L().Info("Shutting down server...")

	// Shutdown with 5 second timeout
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		zap.L().Error("Error during server shutdown", zap.Error(err))
	}

	zap.L().Info("Server gracefully stopped")
}

func httpc() *http.Client {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Transport: otelhttp.NewTransport(transport),
	}

	return httpClient
}

func initTracer(applicationConfig *config.ApplicationConfig) *sdktrace.TracerProvider {
	headers := map[string]string{
		"content-type": "application/json",
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(applicationConfig.OtelTraceEndpoint),
			otlptracehttp.WithHeaders(headers),
			otlptracehttp.WithInsecure(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String("app-go"),
			)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return tp
}
