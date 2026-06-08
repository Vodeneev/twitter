package auth

import (
	"context"
	"strings"
	"time"

	"github.com/Vodeneev/twitter/backend/internal/sqlutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// userColsWithFollow projects the standard user columns plus a `following`
// flag computed against the viewer ($1, nullable).
const userColsWithFollow = `u.id, u.email, u.username, u.display_name, u.bio, u.location, u.website,
	u.avatar_key, u.header_key, u.is_admin, u.is_banned, u.followers_count, u.following_count,
	u.yaps_count, u.email_verified_at, u.created_at,
	COALESCE((SELECT true FROM follows f WHERE f.follower_id = $1 AND f.followee_id = u.id), false)`

func scanUserRow(rows pgx.Rows) (*User, error) {
	var (
		u         User
		avatarKey *string
		headerKey *string
		verified  *time.Time
	)
	if err := rows.Scan(
		&u.ID, &u.Email, &u.Username, &u.DisplayName, &u.Bio, &u.Location, &u.Website,
		&avatarKey, &headerKey, &u.IsAdmin, &u.IsBanned, &u.FollowersCount, &u.FollowingCount,
		&u.YapsCount, &verified, &u.CreatedAt, &u.Following,
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

func (r *UserRepository) collect(ctx context.Context, viewer *uuid.UUID, query string, args ...any) ([]*User, error) {
	full := append([]any{viewer}, args...)
	rows, err := r.pool.Query(ctx, query, full...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*User
	for rows.Next() {
		u, err := scanUserRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

// GetByUsernameWithViewer returns a profile including the viewer's follow state.
func (r *UserRepository) GetByUsernameWithViewer(ctx context.Context, username string, viewer *uuid.UUID) (*User, error) {
	users, err := r.collect(ctx, viewer, `
		SELECT `+userColsWithFollow+`
		FROM users u WHERE u.username = $2`, normalizeUsername(username))
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, ErrUserNotFound
	}
	return users[0], nil
}

func (r *UserRepository) Search(ctx context.Context, q string, viewer *uuid.UUID, limit int) ([]*User, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, nil
	}
	pattern := sqlutil.ContainsPattern(q)
	return r.collect(ctx, viewer, `
		SELECT `+userColsWithFollow+`
		FROM users u
		WHERE u.is_banned = FALSE AND (
			lower(u.username) LIKE $3 ESCAPE '\'
			OR lower(COALESCE(u.display_name, '')) LIKE $3 ESCAPE '\'
		)
		ORDER BY
			CASE
				WHEN lower(u.username) = lower($2) THEN 0
				WHEN lower(u.username) LIKE lower($2) || '%' THEN 1
				WHEN lower(COALESCE(u.display_name, '')) LIKE lower($2) || '%' THEN 2
				WHEN lower(u.username) LIKE $3 ESCAPE '\' THEN 3
				WHEN lower(COALESCE(u.display_name, '')) LIKE $3 ESCAPE '\' THEN 4
				ELSE 5
			END,
			GREATEST(
				word_similarity(lower($2), lower(u.username)),
				word_similarity(lower($2), lower(COALESCE(u.display_name, '')))
			) DESC,
			u.followers_count DESC
		LIMIT $4`, q, pattern, limit)
}

func (r *UserRepository) Followers(ctx context.Context, userID uuid.UUID, viewer *uuid.UUID, limit int) ([]*User, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return r.collect(ctx, viewer, `
		SELECT `+userColsWithFollow+`
		FROM follows fl
		JOIN users u ON u.id = fl.follower_id
		WHERE fl.followee_id = $2
		ORDER BY fl.created_at DESC
		LIMIT $3`, userID, limit)
}

func (r *UserRepository) Following(ctx context.Context, userID uuid.UUID, viewer *uuid.UUID, limit int) ([]*User, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return r.collect(ctx, viewer, `
		SELECT `+userColsWithFollow+`
		FROM follows fl
		JOIN users u ON u.id = fl.followee_id
		WHERE fl.follower_id = $2
		ORDER BY fl.created_at DESC
		LIMIT $3`, userID, limit)
}

// Suggestions returns popular users the viewer does not yet follow ("who to follow").
func (r *UserRepository) Suggestions(ctx context.Context, viewer uuid.UUID, limit int) ([]*User, error) {
	if limit <= 0 || limit > 20 {
		limit = 5
	}
	return r.collect(ctx, &viewer, `
		SELECT `+userColsWithFollow+`
		FROM users u
		WHERE u.id <> $1 AND u.is_banned = FALSE
		  AND NOT EXISTS (SELECT 1 FROM follows f WHERE f.follower_id = $1 AND f.followee_id = u.id)
		ORDER BY u.followers_count DESC, u.created_at DESC
		LIMIT $2`, limit)
}
