package domain

import "context"

type RefreshTokenRepository interface {
	Create(ctx context.Context, refreshToken *RefreshToken) error
}
