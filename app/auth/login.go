package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/knetic0/production-ready-go-cqrs/domain"
	"github.com/knetic0/production-ready-go-cqrs/pkg/config"
	"github.com/knetic0/production-ready-go-cqrs/pkg/security"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}

type LoginHandler struct {
	repository             domain.UserRepository
	refreshTokenRepository domain.RefreshTokenRepository
	config                 config.SecurityConfig
}

func NewLoginHandler(repository domain.UserRepository, refreshTokenRepository domain.RefreshTokenRepository, config config.SecurityConfig) *LoginHandler {
	return &LoginHandler{repository: repository, refreshTokenRepository: refreshTokenRepository, config: config}
}

func (h *LoginHandler) Handle(ctx context.Context, request *LoginRequest) (*LoginResponse, error) {
	user, err := h.repository.GetByEmail(ctx, request.Email)
	if err != nil {
		return nil, err
	}

	if err := security.ValidatePassw(request.Password, user.Password); err != nil {
		return nil, err
	}
	claims := jwt.MapClaims{
		"sub":      user.Id,
		"email":    user.Email,
		"fullName": user.FirstName + " " + user.LastName,
		"exp":      time.Now().Add(time.Minute * time.Duration(h.config.MinutesOfJwtExpiration)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(h.config.JwtSecretKey))
	if err != nil {
		return nil, err
	}

	rt, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshToken := domain.RefreshToken{
		Id:        uuid.New().String(),
		Token:     rt,
		ExpiresAt: time.Now().Add(time.Duration(h.config.HoursOfRefreshTokenExpiration) * time.Hour),
		UserId:    user.Id,
	}

	err = h.refreshTokenRepository.Create(ctx, &refreshToken)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: t, RefreshToken: rt}, nil
}
