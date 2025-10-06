package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/knetic0/production-ready-go-cqrs/domain"
	"github.com/knetic0/production-ready-go-cqrs/pkg/security"
)

type UserCreateRequest struct {
	FirstName string `json:"firstName" validate:"required,min=2"`
	LastName  string `json:"lastName" validate:"required,min=2"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
}

type UserCreateResponse struct{}

type UserCreateHandler struct {
	repository domain.UserRepository
}

func NewUserCreateHandler(repository domain.UserRepository) *UserCreateHandler {
	return &UserCreateHandler{
		repository: repository,
	}
}

func (h *UserCreateHandler) Handle(ctx context.Context, request *UserCreateRequest) (*UserCreateResponse, error) {
	hashed, err := security.HashPassw(request.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Id:        uuid.New().String(),
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		Password:  hashed,
	}

	if err := h.repository.Create(ctx, user); err != nil {
		return nil, err
	}

	return &UserCreateResponse{}, nil
}
