package infrastructure

import (
	"context"

	"github.com/knetic0/production-ready-go-cqrs/domain"
	"gorm.io/gorm"
)

type RefreshTokenRepositoryAdapter struct {
	db *gorm.DB
}

func NewRefreshTokenRepositoryAdapter(db *gorm.DB) *RefreshTokenRepositoryAdapter {
	return &RefreshTokenRepositoryAdapter{db: db}
}

func (r *RefreshTokenRepositoryAdapter) Create(ctx context.Context, refreshToken *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Create(refreshToken).Error
}
