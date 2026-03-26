package domain

import "errors"

var (
	ErrArticleNotFound = errors.New("article not found")
	ErrForbidden       = errors.New("access forbidden")
	ErrInvalidToken    = errors.New("invalid or expired token")
)
