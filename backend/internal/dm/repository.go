package dm

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotParticipant = errors.New("not a conversation participant")
	ErrEmptyMessage   = errors.New("message is empty")
)

const MaxMessageLen = 2000

type KeyToPublicURL func(key string) string

type Participant struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	AvatarURL   string    `json:"avatarUrl"`
}

type Message struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversationId"`
	SenderID       uuid.UUID `json:"senderId"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Conversation struct {
	ID            uuid.UUID   `json:"id"`
	Other         Participant `json:"other"`
	LastMessage   *Message    `json:"lastMessage,omitempty"`
	Unread        int         `json:"unread"`
	LastMessageAt time.Time   `json:"lastMessageAt"`
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

func pairKey(a, b uuid.UUID) string {
	as, bs := a.String(), b.String()
	if as < bs {
		return as + ":" + bs
	}
	return bs + ":" + as
}

// GetOrCreate returns the 1:1 conversation between two users, creating it if needed.
func (r *Repository) GetOrCreate(ctx context.Context, userA, userB uuid.UUID) (uuid.UUID, error) {
	if userA == userB {
		return uuid.Nil, errors.New("cannot DM yourself")
	}
	key := pairKey(userA, userB)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	err = tx.QueryRow(ctx, `SELECT id FROM conversations WHERE pair_key = $1`, key).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := tx.QueryRow(ctx, `INSERT INTO conversations (pair_key) VALUES ($1) RETURNING id`, key).Scan(&id); err != nil {
			return uuid.Nil, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO conversation_participants (conversation_id, user_id) VALUES ($1, $2), ($1, $3)`,
			id, userA, userB); err != nil {
			return uuid.Nil, err
		}
	} else if err != nil {
		return uuid.Nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *Repository) IsParticipant(ctx context.Context, convID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM conversation_participants WHERE conversation_id = $1 AND user_id = $2)`, convID, userID).Scan(&exists)
	return exists, err
}

// Participants returns the two user ids of a conversation.
func (r *Repository) Participants(ctx context.Context, convID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT user_id FROM conversation_participants WHERE conversation_id = $1`, convID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) SendMessage(ctx context.Context, convID, senderID uuid.UUID, body string) (*Message, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, ErrEmptyMessage
	}
	if len([]rune(body)) > MaxMessageLen {
		body = string([]rune(body)[:MaxMessageLen])
	}
	ok, err := r.IsParticipant(ctx, convID, senderID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotParticipant
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	m := &Message{ConversationID: convID, SenderID: senderID, Body: body}
	if err := tx.QueryRow(ctx, `
		INSERT INTO messages (conversation_id, sender_id, body)
		VALUES ($1, $2, $3) RETURNING id, created_at`,
		convID, senderID, body).Scan(&m.ID, &m.CreatedAt); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE conversations SET last_message_at = NOW() WHERE id = $1`, convID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `UPDATE conversation_participants SET last_read_at = NOW() WHERE conversation_id = $1 AND user_id = $2`, convID, senderID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Repository) ListMessages(ctx context.Context, convID, userID uuid.UUID, cursor string, limit int) ([]Message, *string, error) {
	ok, err := r.IsParticipant(ctx, convID, userID)
	if err != nil {
		return nil, nil, err
	}
	if !ok {
		return nil, nil, ErrNotParticipant
	}
	if limit <= 0 || limit > 100 {
		limit = 40
	}
	from := time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
	if t, err := time.Parse(time.RFC3339Nano, cursor); err == nil && cursor != "" {
		from = t
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, conversation_id, sender_id, body, created_at
		FROM messages WHERE conversation_id = $1 AND created_at < $2
		ORDER BY created_at DESC LIMIT $3`, convID, from, limit)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	var items []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderID, &m.Body, &m.CreatedAt); err != nil {
			return nil, nil, err
		}
		items = append(items, m)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	var next *string
	if len(items) == limit && len(items) > 0 {
		s := items[len(items)-1].CreatedAt.UTC().Format(time.RFC3339Nano)
		next = &s
	}
	return items, next, nil
}

func (r *Repository) MarkRead(ctx context.Context, convID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE conversation_participants SET last_read_at = NOW() WHERE conversation_id = $1 AND user_id = $2`, convID, userID)
	return err
}

func (r *Repository) ListConversations(ctx context.Context, userID uuid.UUID) ([]Conversation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.last_message_at,
		       ou.id, ou.username, ou.display_name, ou.avatar_key,
		       (SELECT COUNT(*) FROM messages m
		         WHERE m.conversation_id = c.id AND m.sender_id <> $1 AND m.created_at > cp.last_read_at) AS unread,
		       lm.id, lm.sender_id, lm.body, lm.created_at
		FROM conversation_participants cp
		JOIN conversations c ON c.id = cp.conversation_id
		JOIN conversation_participants op ON op.conversation_id = c.id AND op.user_id <> $1
		JOIN users ou ON ou.id = op.user_id
		LEFT JOIN LATERAL (
		    SELECT id, sender_id, body, created_at FROM messages
		    WHERE conversation_id = c.id ORDER BY created_at DESC LIMIT 1
		) lm ON true
		WHERE cp.user_id = $1
		ORDER BY c.last_message_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Conversation
	for rows.Next() {
		var (
			c         Conversation
			avatarKey *string
			lmID      *uuid.UUID
			lmSender  *uuid.UUID
			lmBody    *string
			lmCreated *time.Time
		)
		if err := rows.Scan(
			&c.ID, &c.LastMessageAt,
			&c.Other.ID, &c.Other.Username, &c.Other.DisplayName, &avatarKey,
			&c.Unread,
			&lmID, &lmSender, &lmBody, &lmCreated,
		); err != nil {
			return nil, err
		}
		if avatarKey != nil {
			c.Other.AvatarURL = r.urlFn(*avatarKey)
		}
		if lmID != nil {
			c.LastMessage = &Message{ID: *lmID, ConversationID: c.ID, SenderID: *lmSender, Body: *lmBody, CreatedAt: *lmCreated}
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
