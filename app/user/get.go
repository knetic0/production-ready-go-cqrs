package user

import (
	"context"

	"github.com/knetic0/production-ready-go-cqrs/domain"
)

type UserGetRequest struct {
	Id string `json:"id" validate:"required,uuid4"`
}

type UserGetResponse struct {
	User *domain.User `json:"user"`
}

type UserGetHandler struct {
	repository domain.UserRepository
}

func NewUserGetHandler(repository domain.UserRepository) *UserGetHandler {
	return &UserGetHandler{repository: repository}
}

func (h *UserGetHandler) Handle(ctx context.Context, request *UserGetRequest) (*UserGetResponse, error) {
	user, err := h.repository.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &UserGetResponse{User: user}, nil
}
