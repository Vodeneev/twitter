package social

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrYapNotFound  = errors.New("yap not found")
	ErrUserNotFound = errors.New("user not found")
	ErrSelfFollow   = errors.New("cannot follow yourself")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func yapAuthor(ctx context.Context, q pgx.Tx, yapID uuid.UUID) (uuid.UUID, error) {
	var author uuid.UUID
	err := q.QueryRow(ctx, `SELECT author_id FROM yaps WHERE id = $1`, yapID).Scan(&author)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrYapNotFound
	}
	return author, err
}

// Like adds a like; returns whether it was newly created and the yap author.
func (r *Repository) Like(ctx context.Context, userID, yapID uuid.UUID) (bool, uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	author, err := yapAuthor(ctx, tx, yapID)
	if err != nil {
		return false, uuid.Nil, err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO likes (user_id, yap_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, yapID)
	if err != nil {
		return false, uuid.Nil, err
	}
	created := tag.RowsAffected() > 0
	if created {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET likes_count = likes_count + 1 WHERE id = $1`, yapID); err != nil {
			return false, uuid.Nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, uuid.Nil, err
	}
	return created, author, nil
}

func (r *Repository) Unlike(ctx context.Context, userID, yapID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM likes WHERE user_id = $1 AND yap_id = $2`, userID, yapID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET likes_count = GREATEST(likes_count - 1, 0) WHERE id = $1`, yapID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) Repost(ctx context.Context, userID, yapID uuid.UUID) (bool, uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	author, err := yapAuthor(ctx, tx, yapID)
	if err != nil {
		return false, uuid.Nil, err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO reposts (user_id, yap_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, yapID)
	if err != nil {
		return false, uuid.Nil, err
	}
	created := tag.RowsAffected() > 0
	if created {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET reposts_count = reposts_count + 1 WHERE id = $1`, yapID); err != nil {
			return false, uuid.Nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, uuid.Nil, err
	}
	return created, author, nil
}

func (r *Repository) Unrepost(ctx context.Context, userID, yapID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM reposts WHERE user_id = $1 AND yap_id = $2`, userID, yapID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET reposts_count = GREATEST(reposts_count - 1, 0) WHERE id = $1`, yapID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) Bookmark(ctx context.Context, userID, yapID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := yapAuthor(ctx, tx, yapID); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `INSERT INTO bookmarks (user_id, yap_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, yapID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET bookmarks_count = bookmarks_count + 1 WHERE id = $1`, yapID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) Unbookmark(ctx context.Context, userID, yapID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM bookmarks WHERE user_id = $1 AND yap_id = $2`, userID, yapID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET bookmarks_count = GREATEST(bookmarks_count - 1, 0) WHERE id = $1`, yapID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// Follow creates an edge follower -> followee; returns whether it was new.
func (r *Repository) Follow(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	if followerID == followeeID {
		return false, ErrSelfFollow
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `INSERT INTO follows (follower_id, followee_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, followerID, followeeID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return false, ErrUserNotFound
		}
		return false, err
	}
	created := tag.RowsAffected() > 0
	if created {
		if _, err := tx.Exec(ctx, `UPDATE users SET following_count = following_count + 1 WHERE id = $1`, followerID); err != nil {
			return false, err
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET followers_count = followers_count + 1 WHERE id = $1`, followeeID); err != nil {
			return false, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return created, nil
}

func (r *Repository) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `DELETE FROM follows WHERE follower_id = $1 AND followee_id = $2`, followerID, followeeID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		if _, err := tx.Exec(ctx, `UPDATE users SET following_count = GREATEST(following_count - 1, 0) WHERE id = $1`, followerID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET followers_count = GREATEST(followers_count - 1, 0) WHERE id = $1`, followeeID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) IsFollowing(ctx context.Context, followerID, followeeID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id = $1 AND followee_id = $2)`, followerID, followeeID).Scan(&exists)
	return exists, err
}
