package domain

import "context"

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Get(ctx context.Context, id string) (*User, error)
	List(ctx context.Context) ([]User, error)
}
