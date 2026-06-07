package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, userID uuid.UUID, ttl time.Duration, rememberMe bool, userAgent, ip string) (*Session, error) {
	expiresAt := time.Now().Add(ttl)
	var ipArg interface{}
	if ip != "" {
		ipArg = ip
	}
	var s Session
	err := r.pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, expires_at, remember_me, user_agent, ip)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, expires_at, remember_me`,
		userID, expiresAt, rememberMe, userAgent, ipArg,
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.RememberMe)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Get returns the session and its user, enforcing expiry.
func (r *SessionRepository) Get(ctx context.Context, id uuid.UUID) (*Session, *User, error) {
	var s Session
	var expiresAt time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT id, user_id, expires_at, remember_me
		FROM sessions WHERE id = $1`, id,
	).Scan(&s.ID, &s.UserID, &expiresAt, &s.RememberMe)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrSessionExpired
		}
		return nil, nil, err
	}
	s.ExpiresAt = expiresAt
	if time.Now().After(expiresAt) {
		_ = r.Delete(ctx, id)
		return nil, nil, ErrSessionExpired
	}

	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, s.UserID)
	u, err := scanUser(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, ErrUserNotFound
		}
		return nil, nil, err
	}
	return &s, u, nil
}

func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, id)
	return err
}

func (r *SessionRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}
