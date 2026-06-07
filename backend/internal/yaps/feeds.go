package yaps

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Vodeneev/twitter/backend/internal/sqlutil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var farFuture = time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)

func cursorOrFuture(cursor string) time.Time {
	if t, ok := decodeCursor(cursor); ok {
		return t
	}
	return farFuture
}

func clampLimit(limit int) int {
	if limit <= 0 || limit > 50 {
		return 20
	}
	return limit
}

// scanFeedRows scans the standard yap projection plus feed context
// (effective_at and an optional reposter), then enriches media/quotes.
func (r *Repository) scanFeedRows(ctx context.Context, viewer *uuid.UUID, rows pgx.Rows) ([]Yap, error) {
	defer rows.Close()
	var out []Yap
	for rows.Next() {
		var (
			y          Yap
			authorID   uuid.UUID
			avatarKey  *string
			username   string
			display    string
			effective  time.Time
			reposterID *uuid.UUID
			rUser      *string
			rDisplay   *string
			rAvatar    *string
		)
		if err := rows.Scan(
			&y.ID, &authorID, &y.Content, &y.ReplyToID, &y.QuoteOfID,
			&y.LikesCount, &y.RepostsCount, &y.RepliesCount, &y.QuotesCount, &y.BookmarksCount, &y.CreatedAt,
			&username, &display, &avatarKey,
			&y.Liked, &y.Reposted, &y.Bookmarked,
			&effective, &reposterID, &rUser, &rDisplay, &rAvatar,
		); err != nil {
			return nil, err
		}
		y.Author = Author{ID: authorID, Username: username, DisplayName: display}
		if avatarKey != nil {
			y.Author.AvatarURL = r.urlFn(*avatarKey)
		}
		y.Media = []Media{}
		if reposterID != nil {
			rb := Author{ID: *reposterID}
			if rUser != nil {
				rb.Username = *rUser
			}
			if rDisplay != nil {
				rb.DisplayName = *rDisplay
			}
			if rAvatar != nil {
				rb.AvatarURL = r.urlFn(*rAvatar)
			}
			y.RepostedBy = &rb
			ea := effective
			y.RepostedAt = &ea
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

const feedTail = `, f.effective_at, ru.id, ru.username, ru.display_name, ru.avatar_key
	FROM feed f
	JOIN yaps y ON y.id = f.yap_id
	JOIN users a ON a.id = y.author_id
	LEFT JOIN users ru ON ru.id = f.reposter_id
	WHERE f.effective_at < $3
	ORDER BY f.effective_at DESC
	LIMIT $4`

// HomeTimeline: top-level yaps and reposts from the viewer and everyone they follow.
func (r *Repository) HomeTimeline(ctx context.Context, viewer uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	rows, err := r.pool.Query(ctx, `
		WITH feed AS (
			SELECT y.id AS yap_id, y.created_at AS effective_at, NULL::uuid AS reposter_id
			FROM yaps y
			WHERE y.reply_to_id IS NULL
			  AND (y.author_id = $2 OR y.author_id IN (SELECT followee_id FROM follows WHERE follower_id = $2))
			UNION ALL
			SELECT rp.yap_id, rp.created_at AS effective_at, rp.user_id AS reposter_id
			FROM reposts rp
			WHERE rp.user_id = $2 OR rp.user_id IN (SELECT followee_id FROM follows WHERE follower_id = $2)
		)
		SELECT `+yapCols+feedTail,
		&viewer, viewer, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	items, err := r.scanFeedRows(ctx, &viewer, rows)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// GlobalTimeline: newest top-level yaps from everyone ("explore").
// Has no owner parameter, so it uses its own placeholder numbering
// ($1 viewer, $2 cursor, $3 limit) instead of the shared feedTail.
func (r *Repository) GlobalTimeline(ctx context.Context, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	rows, err := r.pool.Query(ctx, `
		WITH feed AS (
			SELECT y.id AS yap_id, y.created_at AS effective_at, NULL::uuid AS reposter_id
			FROM yaps y WHERE y.reply_to_id IS NULL
		)
		SELECT `+yapCols+`, f.effective_at, ru.id, ru.username, ru.display_name, ru.avatar_key
		FROM feed f
		JOIN yaps y ON y.id = f.yap_id
		JOIN users a ON a.id = y.author_id
		LEFT JOIN users ru ON ru.id = f.reposter_id
		WHERE f.effective_at < $2
		ORDER BY f.effective_at DESC
		LIMIT $3`,
		viewer, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	items, err := r.scanFeedRows(ctx, viewer, rows)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// UserYaps: a profile's top-level yaps and reposts.
func (r *Repository) UserYaps(ctx context.Context, owner uuid.UUID, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	rows, err := r.pool.Query(ctx, `
		WITH feed AS (
			SELECT y.id AS yap_id, y.created_at AS effective_at, NULL::uuid AS reposter_id
			FROM yaps y WHERE y.author_id = $2 AND y.reply_to_id IS NULL
			UNION ALL
			SELECT rp.yap_id, rp.created_at AS effective_at, rp.user_id AS reposter_id
			FROM reposts rp WHERE rp.user_id = $2
		)
		SELECT `+yapCols+feedTail,
		viewer, owner, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	items, err := r.scanFeedRows(ctx, viewer, rows)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// UserReplies: a profile's replies.
func (r *Repository) UserReplies(ctx context.Context, owner uuid.UUID, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	rows, err := r.pool.Query(ctx, `
		WITH feed AS (
			SELECT y.id AS yap_id, y.created_at AS effective_at, NULL::uuid AS reposter_id
			FROM yaps y WHERE y.author_id = $2 AND y.reply_to_id IS NOT NULL
		)
		SELECT `+yapCols+feedTail,
		viewer, owner, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	items, err := r.scanFeedRows(ctx, viewer, rows)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// UserMedia: a profile's yaps that include media.
func (r *Repository) UserMedia(ctx context.Context, owner uuid.UUID, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	rows, err := r.pool.Query(ctx, `
		WITH feed AS (
			SELECT y.id AS yap_id, y.created_at AS effective_at, NULL::uuid AS reposter_id
			FROM yaps y
			WHERE y.author_id = $2 AND EXISTS (SELECT 1 FROM yap_media m WHERE m.yap_id = y.id)
		)
		SELECT `+yapCols+feedTail,
		viewer, owner, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	items, err := r.scanFeedRows(ctx, viewer, rows)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// UserLikes: yaps the profile owner has liked.
func (r *Repository) UserLikes(ctx context.Context, owner uuid.UUID, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	rows, err := r.pool.Query(ctx, `
		WITH feed AS (
			SELECT l.yap_id AS yap_id, l.created_at AS effective_at, NULL::uuid AS reposter_id
			FROM likes l WHERE l.user_id = $2
		)
		SELECT `+yapCols+feedTail,
		viewer, owner, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	items, err := r.scanFeedRows(ctx, viewer, rows)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// Bookmarks: yaps the viewer bookmarked.
func (r *Repository) Bookmarks(ctx context.Context, viewer uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	rows, err := r.pool.Query(ctx, `
		WITH feed AS (
			SELECT b.yap_id AS yap_id, b.created_at AS effective_at, NULL::uuid AS reposter_id
			FROM bookmarks b WHERE b.user_id = $2
		)
		SELECT `+yapCols+feedTail,
		&viewer, viewer, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	items, err := r.scanFeedRows(ctx, &viewer, rows)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// Replies returns direct replies to a yap, oldest first.
func (r *Repository) Replies(ctx context.Context, yapID uuid.UUID, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	from := time.Time{}
	if t, ok := decodeCursor(cursor); ok {
		from = t
	}
	items, err := r.query(ctx, viewer, `
		SELECT `+yapCols+`
		FROM yaps y JOIN users a ON a.id = y.author_id
		WHERE y.reply_to_id = $2 AND y.created_at > $3
		ORDER BY y.created_at ASC
		LIMIT $4`, yapID, from, limit)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// Ancestors walks the parent chain of a yap, root-first (for thread context).
func (r *Repository) Ancestors(ctx context.Context, yapID uuid.UUID, viewer *uuid.UUID) ([]Yap, error) {
	items, err := r.query(ctx, viewer, `
		WITH RECURSIVE chain AS (
			SELECT id, reply_to_id, 0 AS depth FROM yaps WHERE id = $2
			UNION ALL
			SELECT y.id, y.reply_to_id, c.depth + 1
			FROM yaps y JOIN chain c ON y.id = c.reply_to_id
		)
		SELECT `+yapCols+`
		FROM chain ch
		JOIN yaps y ON y.id = ch.id
		JOIN users a ON a.id = y.author_id
		WHERE ch.id <> $2
		ORDER BY ch.depth DESC`, yapID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// HashtagTimeline: yaps tagged with a hashtag.
func (r *Repository) HashtagTimeline(ctx context.Context, tag string, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	items, err := r.query(ctx, viewer, `
		SELECT `+yapCols+`
		FROM yaps y
		JOIN users a ON a.id = y.author_id
		JOIN yap_hashtags yh ON yh.yap_id = y.id
		JOIN hashtags h ON h.id = yh.hashtag_id
		WHERE h.tag = $2 AND y.created_at < $3
		ORDER BY y.created_at DESC
		LIMIT $4`, tag, cursorOrFuture(cursor), limit)
	if err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextCursor(items)}, nil
}

// SearchYaps matches substrings and ranks by relevance (exact > prefix > contains > trigram).
func (r *Repository) SearchYaps(ctx context.Context, q string, viewer *uuid.UUID, cursor string, limit int) (Page, error) {
	limit = clampLimit(limit)
	q = strings.TrimSpace(q)
	if q == "" {
		return Page{}, nil
	}
	offset := decodeSearchOffset(cursor)
	pattern := sqlutil.ContainsPattern(q)
	rows, err := r.pool.Query(ctx, `
		WITH ranked AS (
			SELECT y.id,
				(
					CASE
						WHEN lower(y.content) = lower($2) THEN 1000.0
						WHEN lower(y.content) LIKE lower($2) || '%' THEN 800.0
						WHEN lower(y.content) LIKE $3 ESCAPE '\' THEN 500.0
						ELSE 0.0
					END
					+ GREATEST(
						word_similarity(lower($2), lower(y.content)),
						similarity(lower(y.content), lower($2))
					) * 200.0
				) AS score,
				y.created_at
			FROM yaps y
			WHERE lower(y.content) LIKE $3 ESCAPE '\'
			   OR word_similarity(lower($2), lower(y.content)) > 0.2
			   OR similarity(lower(y.content), lower($2)) > 0.12
		)
		SELECT `+yapCols+`
		FROM ranked r
		JOIN yaps y ON y.id = r.id
		JOIN users a ON a.id = y.author_id
		ORDER BY r.score DESC, r.created_at DESC
		OFFSET $4 LIMIT $5`,
		viewer, q, pattern, offset, limit)
	if err != nil {
		return Page{}, err
	}
	defer rows.Close()
	var items []Yap
	for rows.Next() {
		y, err := r.scanYap(rows)
		if err != nil {
			return Page{}, err
		}
		items = append(items, y)
	}
	if err := rows.Err(); err != nil {
		return Page{}, err
	}
	if err := r.enrich(ctx, viewer, items); err != nil {
		return Page{}, err
	}
	return Page{Items: items, NextCursor: nextSearchOffsetCursor(offset, len(items), limit)}, nil
}

func decodeSearchOffset(cursor string) int {
	if !strings.HasPrefix(cursor, "s:") {
		return 0
	}
	n, err := strconv.Atoi(cursor[2:])
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func nextSearchOffsetCursor(offset, count, limit int) *string {
	if count < limit {
		return nil
	}
	s := fmt.Sprintf("s:%d", offset+count)
	return &s
}
