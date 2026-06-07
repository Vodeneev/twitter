package notifications

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// KeyToPublicURL resolves a storage key to a browser-facing URL.
type KeyToPublicURL func(key string) string

type Actor struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   string    `json:"avatarUrl"`
}

type Notification struct {
	ID         uuid.UUID  `json:"id"`
	Type       string     `json:"type"`
	Actor      Actor      `json:"actor"`
	YapID      *uuid.UUID `json:"yapId,omitempty"`
	YapPreview string     `json:"yapPreview,omitempty"`
	Read       bool       `json:"read"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type Page struct {
	Items      []Notification `json:"items"`
	NextCursor *string        `json:"nextCursor,omitempty"`
}

type Repository struct {
	pool  *pgxpool.Pool
	urlFn KeyToPublicURL
}

func NewRepository(pool *pgxpool.Pool, urlFn KeyToPublicURL) *Repository {
	if urlFn == nil {
		urlFn = func(string) string { return "" }
	}
	return &Repository{pool: pool, urlFn: urlFn}
}

// Create inserts a notification. A no-op when recipient == actor.
// Returns the assembled notification (nil if skipped) for realtime delivery.
func (r *Repository) Create(ctx context.Context, userID, actorID uuid.UUID, typ string, yapID *uuid.UUID) (*Notification, error) {
	if userID == actorID {
		return nil, nil
	}
	var id uuid.UUID
	var createdAt time.Time
	err := r.pool.QueryRow(ctx, `
		INSERT INTO notifications (user_id, actor_id, type, yap_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`,
		userID, actorID, typ, yapID,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	n := &Notification{ID: id, Type: typ, YapID: yapID, CreatedAt: createdAt}
	var avatarKey *string
	err = r.pool.QueryRow(ctx, `SELECT id, username, display_name, avatar_key FROM users WHERE id = $1`, actorID).
		Scan(&n.Actor.ID, &n.Actor.Username, &n.Actor.DisplayName, &avatarKey)
	if err != nil {
		return nil, err
	}
	if avatarKey != nil {
		n.Actor.AvatarURL = r.urlFn(*avatarKey)
	}
	if yapID != nil {
		var content string
		if err := r.pool.QueryRow(ctx, `SELECT content FROM yaps WHERE id = $1`, *yapID).Scan(&content); err == nil {
			n.YapPreview = preview(content)
		}
	}
	return n, nil
}

func preview(s string) string {
	r := []rune(s)
	if len(r) > 80 {
		return string(r[:80]) + "…"
	}
	return s
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, cursor string, limit int) (Page, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	from := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	if t, err := time.Parse(time.RFC3339Nano, cursor); err == nil && cursor != "" {
		from = t
	}
	rows, err := r.pool.Query(ctx, `
		SELECT n.id, n.type, n.yap_id, n.read_at, n.created_at,
		       a.id, a.username, a.display_name, a.avatar_key,
		       y.content
		FROM notifications n
		JOIN users a ON a.id = n.actor_id
		LEFT JOIN yaps y ON y.id = n.yap_id
		WHERE n.user_id = $1 AND n.created_at < $2
		ORDER BY n.created_at DESC
		LIMIT $3`, userID, from, limit)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()

	var items []Notification
	for rows.Next() {
		var (
			n         Notification
			readAt    *time.Time
			avatarKey *string
			content   *string
		)
		if err := rows.Scan(
			&n.ID, &n.Type, &n.YapID, &readAt, &n.CreatedAt,
			&n.Actor.ID, &n.Actor.Username, &n.Actor.DisplayName, &avatarKey,
			&content,
		); err != nil {
			return Page{}, err
		}
		n.Read = readAt != nil
		if avatarKey != nil {
			n.Actor.AvatarURL = r.urlFn(*avatarKey)
		}
		if content != nil {
			n.YapPreview = preview(*content)
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return Page{}, err
	}
	var next *string
	if len(items) >= limit {
		s := items[len(items)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
		next = &s
	}
	return Page{Items: items, NextCursor: next}, nil
}

func (r *Repository) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&n)
	return n, err
}

func (r *Repository) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL`, userID)
	return err
}
