package infrastructure

import (
	"context"
	"fmt"

	"github.com/knetic0/production-ready-go-cqrs/domain"
	"gorm.io/gorm"
)

type UserRepositoryAdapter struct {
	db *gorm.DB
}

func NewUserRepositoryAdapter(db *gorm.DB) *UserRepositoryAdapter {
	return &UserRepositoryAdapter{db: db}
}

func (r *UserRepositoryAdapter) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepositoryAdapter) Get(ctx context.Context, id string) (*domain.User, error) {
	return r.getByField(ctx, "id", id)
}

func (r *UserRepositoryAdapter) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.getByField(ctx, "email", email)
}

func (r *UserRepositoryAdapter) List(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepositoryAdapter) getByField(ctx context.Context, field string, value any) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).Where(fmt.Sprintf("%s = ?", field), value).Take(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
