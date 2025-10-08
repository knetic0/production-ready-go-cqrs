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

type MeHandler struct{}

func NewMeHandler() *MeHandler {
	return &MeHandler{}
}

func (h *MeHandler) Handle(ctx context.Context, request *MeRequest) (*MeResponse, error) {
	user, ok := ctx.Value("user").(*domain.User)
	if !ok {
		return nil, ErrUnauthorized
	}

	return &MeResponse{User: user}, nil
}
