package ports

import "context"

type TokenValidator interface {
	Validate(ctx context.Context, token string) (string, error)
}
