package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

const userColumns = `id, email, username, display_name, bio, location, website,
	avatar_key, header_key, is_admin, is_banned, followers_count, following_count,
	yaps_count, email_verified_at, created_at`

func scanUser(row pgx.Row) (*User, error) {
	var (
		u         User
		avatarKey *string
		headerKey *string
		verified  *time.Time
	)
	if err := row.Scan(
		&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.Bio, &u.Location, &u.Website,
		&avatarKey, &headerKey, &u.IsAdmin, &u.IsBanned, &u.FollowersCount, &u.FollowingCount,
		&u.YapsCount, &verified, &u.CreatedAt,
	); err != nil {
		return nil, err
	}
	if avatarKey != nil {
		u.AvatarKey = *avatarKey
	}
	if headerKey != nil {
		u.HeaderKey = *headerKey
	}
	u.EmailVerified = verified != nil
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, username, email, passwordHash, displayName string) (*User, error) {
	if strings.TrimSpace(displayName) == "" {
		displayName = username
	}
	row := r.pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING `+userColumns,
		username, email, passwordHash, displayName,
	)
	u, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "email") {
				return nil, ErrEmailTaken
			}
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE username = $1`, normalizeUsername(username))
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, NormalizeEmail(email))
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

// GetByLogin accepts an email or a username and returns the user plus password hash.
func (r *UserRepository) GetByLogin(ctx context.Context, identifier string) (*User, string, error) {
	var hash string
	var (
		u         User
		avatarKey *string
		headerKey *string
		verified  *time.Time
	)
	row := r.pool.QueryRow(ctx, `
		SELECT `+userColumns+`, password_hash
		FROM users WHERE email = $1 OR username = $1`,
		strings.ToLower(strings.TrimSpace(identifier)),
	)
	if err := row.Scan(
		&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.Bio, &u.Location, &u.Website,
		&avatarKey, &headerKey, &u.IsAdmin, &u.IsBanned, &u.FollowersCount, &u.FollowingCount,
		&u.YapsCount, &verified, &u.CreatedAt, &hash,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}
	if avatarKey != nil {
		u.AvatarKey = *avatarKey
	}
	if headerKey != nil {
		u.HeaderKey = *headerKey
	}
	u.EmailVerified = verified != nil
	return &u, hash, nil
}

// UpdateProfileInput carries optional profile fields; nil means "leave unchanged".
type UpdateProfileInput struct {
	DisplayName *string
	Bio         *string
	Location    *string
	Website     *string
	AvatarKey   *string
	HeaderKey   *string
}

func (r *UserRepository) UpdateProfile(ctx context.Context, id uuid.UUID, in UpdateProfileInput) (*User, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE users SET
			display_name = COALESCE($2, display_name),
			bio          = COALESCE($3, bio),
			location     = COALESCE($4, location),
			website      = COALESCE($5, website),
			avatar_key   = COALESCE($6, avatar_key),
			header_key   = COALESCE($7, header_key),
			updated_at   = NOW()
		WHERE id = $1
		RETURNING `+userColumns,
		id, in.DisplayName, in.Bio, in.Location, in.Website, in.AvatarKey, in.HeaderKey,
	)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) GrantAdminByEmails(ctx context.Context, emails []string) (int64, error) {
	if len(emails) == 0 {
		return 0, nil
	}
	normalized := make([]string, 0, len(emails))
	for _, e := range emails {
		normalized = append(normalized, NormalizeEmail(e))
	}
	tag, err := r.pool.Exec(ctx, `UPDATE users SET is_admin = TRUE WHERE email = ANY($1)`, normalized)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
