package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const passwordResetTTL = time.Hour

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Issue(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, hash, err := newToken()
	if err != nil {
		return "", err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO password_resets (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`,
		userID, hash, time.Now().Add(passwordResetTTL),
	)
	if err != nil {
		return "", err
	}
	return raw, nil
}

// ConsumeAndSetPassword validates the token, sets the new hash, deletes the token,
// and revokes all sessions for the user — all atomically.
func (r *PasswordResetRepository) ConsumeAndSetPassword(ctx context.Context, rawToken, newHash string) error {
	hash := hashToken(rawToken)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	var consumedAt *time.Time
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id, consumed_at, expires_at
		FROM password_resets WHERE token_hash = $1
		FOR UPDATE`, hash,
	).Scan(&userID, &consumedAt, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPasswordResetInvalid
		}
		return err
	}
	if consumedAt != nil || time.Now().After(expiresAt) {
		return ErrPasswordResetInvalid
	}

	if _, err := tx.Exec(ctx, `UPDATE password_resets SET consumed_at = NOW() WHERE token_hash = $1`, hash); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`, userID, newHash); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
