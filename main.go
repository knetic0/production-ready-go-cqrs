package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/knetic0/production-ready-go-cqrs/app/auth"
	"github.com/knetic0/production-ready-go-cqrs/app/healthcheck"
	"github.com/knetic0/production-ready-go-cqrs/app/user"
	"github.com/knetic0/production-ready-go-cqrs/infrastructure"
	"github.com/knetic0/production-ready-go-cqrs/pkg/config"
)

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

	app := fiber.New(fiber.Config{
		IdleTimeout:  5 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Concurrency:  256 * 1024,
	})

	db := infrastructure.NewPostgreAdapter(applicationConfig.Postgre.DSN)
	healthCheckHandler := healthcheck.NewHealthCheckHandler()
	userRepository := infrastructure.NewUserRepositoryAdapter(db)
	userCreateHandler := user.NewUserCreateHandler(userRepository)
	userGetHandler := user.NewUserGetHandler(userRepository)
	userListHandler := user.NewUserListHandler(userRepository)
	loginHandler := auth.NewLoginHandler(userRepository, applicationConfig.Security)

	app.Post("/login/", handle[auth.LoginRequest, auth.LoginResponse](loginHandler))

	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{
			JWTAlg: jwtware.HS256,
			Key:    []byte(applicationConfig.Security.JwtSecretKey),
		},
	}))

	app.Get("/healthcheck", handle[healthcheck.HealthCheckRequest, healthcheck.HealthCheckResponse](healthCheckHandler))
	app.Post("/users/", handle[user.UserCreateRequest, user.UserCreateResponse](userCreateHandler))
	app.Get("/users/:id", handle[user.UserGetRequest, user.UserGetResponse](userGetHandler))
	app.Get("/users/", handle[user.UserListRequest, user.UserListResponse](userListHandler))

	if err := app.Listen(fmt.Sprintf("0.0.0.0:%s", applicationConfig.Server.Port)); err != nil {
		os.Exit(1)
	}
}
