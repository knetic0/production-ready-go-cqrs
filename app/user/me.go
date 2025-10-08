package user

import (
	"context"
	"errors"

	"github.com/knetic0/production-ready-go-cqrs/domain"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type MeRequest struct{}

type MeResponse struct {
	User *domain.User `json:"user"`
}

type MeHandler struct {
	repository domain.UserRepository
}

func NewMeHandler(repository domain.UserRepository) *MeHandler {
	return &MeHandler{
		repository: repository,
	}
}

func (h *MeHandler) Handle(ctx context.Context, request *MeRequest) (*MeResponse, error) {
	userId, ok := ctx.Value("userId").(string)
	if !ok {
		return nil, ErrUnauthorized
	}

	user, err := h.repository.Get(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &MeResponse{User: user}, nil
}
