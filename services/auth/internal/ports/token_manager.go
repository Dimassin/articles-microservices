package ports

import (
	"auth/internal/domain"
	"context"
)

type TokenManager interface {
	GenerateAccessToken(ctx context.Context, user *domain.User) (string, error)

	GenerateRefreshToken(ctx context.Context, user *domain.User) (string, error)

	VerifyToken(ctx context.Context, token string) (*domain.UserClaims, error)
}
