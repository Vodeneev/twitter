package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const verificationTTL = 24 * time.Hour

type VerificationRepository struct {
	pool *pgxpool.Pool
}

func NewVerificationRepository(pool *pgxpool.Pool) *VerificationRepository {
	return &VerificationRepository{pool: pool}
}

// Issue creates a new single-use token and returns the raw (un-hashed) value.
func (r *VerificationRepository) Issue(ctx context.Context, userID uuid.UUID) (string, error) {
	raw, hash, err := newToken()
	if err != nil {
		return "", err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO email_verifications (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`,
		userID, hash, time.Now().Add(verificationTTL),
	)
	if err != nil {
		return "", err
	}
	return raw, nil
}

// Consume validates a raw token, marks the user verified, and returns the user.
func (r *VerificationRepository) Consume(ctx context.Context, rawToken string) (*User, error) {
	hash := hashToken(rawToken)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var userID uuid.UUID
	var consumedAt *time.Time
	var expiresAt time.Time
	err = tx.QueryRow(ctx, `
		SELECT user_id, consumed_at, expires_at
		FROM email_verifications WHERE token_hash = $1
		FOR UPDATE`, hash,
	).Scan(&userID, &consumedAt, &expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrVerificationInvalid
		}
		return nil, err
	}
	if consumedAt != nil {
		// Token already used: still return the (already verified) user idempotently.
		row := tx.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, userID)
		u, scanErr := scanUser(row)
		if scanErr != nil {
			return nil, ErrVerificationInvalid
		}
		_ = tx.Commit(ctx)
		return u, ErrAlreadyVerified
	}
	if time.Now().After(expiresAt) {
		return nil, ErrVerificationInvalid
	}

	if _, err := tx.Exec(ctx, `UPDATE email_verifications SET consumed_at = NOW() WHERE token_hash = $1`, hash); err != nil {
		return nil, err
	}
	row := tx.QueryRow(ctx, `
		UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()), updated_at = NOW()
		WHERE id = $1
		RETURNING `+userColumns, userID)
	u, err := scanUser(row)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return u, nil
}

func newToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
