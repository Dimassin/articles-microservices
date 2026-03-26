package auth

import (
	"blog/internal/domain"
	"context"

	"github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
	secretKey string
}

func NewJWTValidator(secretKey string) *JWTValidator {
	return &JWTValidator{
		secretKey: secretKey,
	}
}

func (v *JWTValidator) Validate(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(v.secretKey), nil
	})

	if err != nil {
		return "", domain.ErrInvalidToken
	}

	if !token.Valid {
		return "", domain.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", domain.ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", domain.ErrInvalidToken
	}

	return userID, nil
}
