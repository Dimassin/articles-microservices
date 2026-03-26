package auth

import "context"

type MockTokenValidator struct{}

func NewMockTokenValidator() *MockTokenValidator {
	return &MockTokenValidator{}
}

func (v *MockTokenValidator) Validate(ctx context.Context, token string) (string, error) {
	return "test-user-id", nil
}
