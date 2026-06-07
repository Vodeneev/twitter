package yaps

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

// yapCols is the standard projection. $1 is always the viewer id (nullable).
const yapCols = `y.id, y.author_id, y.content, y.reply_to_id, y.quote_of_id,
	y.likes_count, y.reposts_count, y.replies_count, y.quotes_count, y.bookmarks_count, y.created_at,
	a.username, a.display_name, a.avatar_key,
	COALESCE((SELECT true FROM likes l WHERE l.yap_id = y.id AND l.user_id = $1), false),
	COALESCE((SELECT true FROM reposts rp WHERE rp.yap_id = y.id AND rp.user_id = $1), false),
	COALESCE((SELECT true FROM bookmarks bm WHERE bm.yap_id = y.id AND bm.user_id = $1), false)`

func (r *Repository) scanYap(rows pgx.Rows) (Yap, error) {
	var (
		y         Yap
		authorID  uuid.UUID
		avatarKey *string
		username  string
		display   string
	)
	if err := rows.Scan(
		&y.ID, &authorID, &y.Content, &y.ReplyToID, &y.QuoteOfID,
		&y.LikesCount, &y.RepostsCount, &y.RepliesCount, &y.QuotesCount, &y.BookmarksCount, &y.CreatedAt,
		&username, &display, &avatarKey,
		&y.Liked, &y.Reposted, &y.Bookmarked,
	); err != nil {
		return Yap{}, err
	}
	y.Author = Author{ID: authorID, Username: username, DisplayName: display}
	if avatarKey != nil {
		y.Author.AvatarURL = r.urlFn(*avatarKey)
	}
	y.Media = []Media{}
	return y, nil
}

// query runs a yap SELECT (viewer is $1, extra placeholders start at $2) and enriches results.
func (r *Repository) query(ctx context.Context, viewer *uuid.UUID, query string, extraArgs ...any) ([]Yap, error) {
	args := append([]any{viewer}, extraArgs...)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Yap
	for rows.Next() {
		y, err := r.scanYap(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, y)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.enrich(ctx, viewer, out); err != nil {
		return nil, err
	}
	return out, nil
}

// enrich attaches media to all yaps and loads quoted yaps one level deep.
func (r *Repository) enrich(ctx context.Context, viewer *uuid.UUID, items []Yap) error {
	if len(items) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, 0, len(items))
	quoteIDs := make([]uuid.UUID, 0)
	for i := range items {
		ids = append(ids, items[i].ID)
		if items[i].QuoteOfID != nil {
			quoteIDs = append(quoteIDs, *items[i].QuoteOfID)
		}
	}
	if err := r.attachMedia(ctx, ids, items); err != nil {
		return err
	}
	if len(quoteIDs) > 0 {
		quotes, err := r.byIDs(ctx, viewer, quoteIDs)
		if err != nil {
			return err
		}
		for i := range items {
			if items[i].QuoteOfID != nil {
				if q, ok := quotes[*items[i].QuoteOfID]; ok {
					qc := q
					items[i].QuoteOf = &qc
				}
			}
		}
	}
	return nil
}

func (r *Repository) attachMedia(ctx context.Context, ids []uuid.UUID, items []Yap) error {
	rows, err := r.pool.Query(ctx, `
		SELECT yap_id, id, s3_key, position
		FROM yap_media WHERE yap_id = ANY($1)
		ORDER BY yap_id, position`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	byYap := map[uuid.UUID][]Media{}
	for rows.Next() {
		var yapID, mediaID uuid.UUID
		var key string
		var pos int
		if err := rows.Scan(&yapID, &mediaID, &key, &pos); err != nil {
			return err
		}
		byYap[yapID] = append(byYap[yapID], Media{ID: mediaID, URL: r.urlFn(key), Position: pos})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for i := range items {
		if m, ok := byYap[items[i].ID]; ok {
			items[i].Media = m
		}
	}
	return nil
}

// byIDs loads a set of yaps (without recursive quote expansion) keyed by id.
func (r *Repository) byIDs(ctx context.Context, viewer *uuid.UUID, ids []uuid.UUID) (map[uuid.UUID]Yap, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+yapCols+`
		FROM yaps y JOIN users a ON a.id = y.author_id
		WHERE y.id = ANY($2)`, viewer, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[uuid.UUID]Yap{}
	var list []Yap
	for rows.Next() {
		y, err := r.scanYap(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, y)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err := r.attachMedia(ctx, ids, list); err != nil {
		return nil, err
	}
	for _, y := range list {
		result[y.ID] = y
	}
	return result, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID, viewer *uuid.UUID) (*Yap, error) {
	items, err := r.query(ctx, viewer, `
		SELECT `+yapCols+`
		FROM yaps y JOIN users a ON a.id = y.author_id
		WHERE y.id = $2`, id)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, ErrNotFound
	}
	return &items[0], nil
}

// Create inserts a yap with media, hashtags and mentions, bumps counters,
// and returns the assembled yap plus the set of mentioned user ids.
func (r *Repository) Create(ctx context.Context, in CreateInput) (*Yap, []uuid.UUID, error) {
	content := strings.TrimSpace(in.Content)
	if len([]rune(content)) > MaxContentLen {
		return nil, nil, ErrTooLong
	}
	if content == "" && len(in.MediaKeys) == 0 {
		return nil, nil, ErrEmpty
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO yaps (author_id, content, reply_to_id, quote_of_id, search_vector)
		VALUES ($1, $2, $3, $4, to_tsvector('simple', $2))
		RETURNING id`,
		in.AuthorID, content, in.ReplyToID, in.QuoteOfID,
	).Scan(&id)
	if err != nil {
		return nil, nil, err
	}

	for i, key := range in.MediaKeys {
		if _, err := tx.Exec(ctx, `INSERT INTO yap_media (yap_id, s3_key, position) VALUES ($1, $2, $3)`, id, key, i); err != nil {
			return nil, nil, err
		}
	}

	if _, err := tx.Exec(ctx, `UPDATE users SET yaps_count = yaps_count + 1 WHERE id = $1`, in.AuthorID); err != nil {
		return nil, nil, err
	}
	if in.ReplyToID != nil {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET replies_count = replies_count + 1 WHERE id = $1`, *in.ReplyToID); err != nil {
			return nil, nil, err
		}
	}
	if in.QuoteOfID != nil {
		if _, err := tx.Exec(ctx, `UPDATE yaps SET quotes_count = quotes_count + 1 WHERE id = $1`, *in.QuoteOfID); err != nil {
			return nil, nil, err
		}
	}

	if err := upsertHashtags(ctx, tx, id, parseHashtags(content)); err != nil {
		return nil, nil, err
	}
	mentionedIDs, err := upsertMentions(ctx, tx, id, parseMentions(content))
	if err != nil {
		return nil, nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	y, err := r.GetByID(ctx, id, &in.AuthorID)
	if err != nil {
		return nil, nil, err
	}
	return y, mentionedIDs, nil
}

// Delete removes a yap owned by authorID and decrements counters.
func (r *Repository) Delete(ctx context.Context, id, authorID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var owner uuid.UUID
	var replyTo, quoteOf *uuid.UUID
	err = tx.QueryRow(ctx, `SELECT author_id, reply_to_id, quote_of_id FROM yaps WHERE id = $1`, id).Scan(&owner, &replyTo, &quoteOf)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	if owner != authorID {
		return ErrForbidden
	}

	if _, err := tx.Exec(ctx, `DELETE FROM yaps WHERE id = $1`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET yaps_count = GREATEST(yaps_count - 1, 0) WHERE id = $1`, authorID); err != nil {
		return err
	}
	if replyTo != nil {
		_, _ = tx.Exec(ctx, `UPDATE yaps SET replies_count = GREATEST(replies_count - 1, 0) WHERE id = $1`, *replyTo)
	}
	if quoteOf != nil {
		_, _ = tx.Exec(ctx, `UPDATE yaps SET quotes_count = GREATEST(quotes_count - 1, 0) WHERE id = $1`, *quoteOf)
	}
	return tx.Commit(ctx)
}

// --- cursor helpers -------------------------------------------------------

func decodeCursor(cursor string) (time.Time, bool) {
	if cursor == "" {
		return time.Time{}, false
	}
	t, err := time.Parse(time.RFC3339Nano, cursor)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func nextCursor(items []Yap, limit int) *string {
	if len(items) == 0 || len(items) < limit {
		return nil
	}
	last := items[len(items)-1]
	t := last.CreatedAt
	if last.RepostedAt != nil {
		t = *last.RepostedAt
	}
	s := t.UTC().Format(time.RFC3339Nano)
	return &s
}
