package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/knetic0/production-ready-go-cqrs/domain"
	"github.com/knetic0/production-ready-go-cqrs/pkg/config"
	"github.com/knetic0/production-ready-go-cqrs/pkg/security"
)

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type LoginHandler struct {
	repository domain.UserRepository
	config     config.SecurityConfig
}

func NewLoginHandler(repository domain.UserRepository, config config.SecurityConfig) *LoginHandler {
	return &LoginHandler{repository: repository, config: config}
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
		"email":    user.Email,
		"fullName": user.FirstName + " " + user.LastName,
		"exp":      time.Now().Add(time.Minute * time.Duration(h.config.MinutesOfExpiration)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(h.config.JwtSecretKey))
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: t}, nil
}
