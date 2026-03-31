package postgres

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"errors"
	"time"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{
		db: db,
	}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	return err
}

func (r *RefreshTokenRepository) FindByToken(ctx context.Context, token string) (string, error) {
	var userID string
	var revoked bool
	var expiresAt time.Time

	query := `SELECT user_id, revoked, expires_at FROM refresh_tokens WHERE token = $1`
	err := r.db.QueryRowContext(ctx, query, token).Scan(&userID, &revoked, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.ErrInvalidToken
	}
	if err != nil {
		return "", err
	}
	if revoked {
		return "", domain.ErrInvalidToken
	}
	if time.Now().After(expiresAt) {
		return "", domain.ErrInvalidToken
	}
	return userID, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrInvalidToken
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
