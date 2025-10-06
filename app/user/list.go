package user

import (
	"context"

	"github.com/knetic0/production-ready-go-cqrs/domain"
)

type UserListRequest struct{}

type UserListResponse struct {
	Users []domain.User `json:"users"`
}

type UserListHandler struct {
	repository domain.UserRepository
}

func NewUserListHandler(repository domain.UserRepository) *UserListHandler {
	return &UserListHandler{repository: repository}
}

func (h *UserListHandler) Handle(ctx context.Context, request *UserListRequest) (*UserListResponse, error) {
	users, err := h.repository.List(ctx)
	if err != nil {
		return nil, err
	}
	return &UserListResponse{Users: users}, nil
}
