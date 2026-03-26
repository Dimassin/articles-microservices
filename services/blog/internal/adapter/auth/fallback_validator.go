package auth

import (
	"blog/internal/domain"
	"context"
	"log"
)

type FallbackValidator struct {
	local *JWTValidator
	grpc  *GRPCTokenValidator
}

func NewFallbackValidator(local *JWTValidator, grpc *GRPCTokenValidator) *FallbackValidator {
	return &FallbackValidator{
		local: local,
		grpc:  grpc,
	}
}

func (v *FallbackValidator) Validate(ctx context.Context, token string) (string, error) {
	// 1. Пробуем локальную проверку JWT
	userID, err := v.local.Validate(ctx, token)
	if err == nil {
		log.Println("[AUTH] ✅ Local validation SUCCESS")
		return userID, nil
	}

	// 2. Если локально не прошло — идем в auth через gRPC
	userID, err = v.grpc.Validate(ctx, token)
	if err == nil {
		log.Println("[AUTH] ✅ gRPC validation SUCCESS")
		return userID, nil
	}

	log.Println("[AUTH] ❌ Both methods failed")
	return "", domain.ErrInvalidToken
}
