package yaps

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	hashtagRe = regexp.MustCompile(`#([\p{L}\p{N}_]{1,50})`)
	mentionRe = regexp.MustCompile(`@([a-zA-Z0-9_]{1,20})`)
)

func parseHashtags(content string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range hashtagRe.FindAllStringSubmatch(content, -1) {
		tag := strings.ToLower(m[1])
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func parseMentions(content string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range mentionRe.FindAllStringSubmatch(content, -1) {
		u := strings.ToLower(m[1])
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

func upsertHashtags(ctx context.Context, tx pgx.Tx, yapID uuid.UUID, tags []string) error {
	for _, tag := range tags {
		var hashtagID int64
		if err := tx.QueryRow(ctx, `
			INSERT INTO hashtags (tag) VALUES ($1)
			ON CONFLICT (tag) DO UPDATE SET tag = EXCLUDED.tag
			RETURNING id`, tag).Scan(&hashtagID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO yap_hashtags (yap_id, hashtag_id) VALUES ($1, $2)
			ON CONFLICT DO NOTHING`, yapID, hashtagID); err != nil {
			return err
		}
	}
	return nil
}

// upsertMentions resolves usernames to ids, stores mention rows, and returns the ids.
func upsertMentions(ctx context.Context, tx pgx.Tx, yapID uuid.UUID, usernames []string) ([]uuid.UUID, error) {
	if len(usernames) == 0 {
		return nil, nil
	}
	rows, err := tx.Query(ctx, `SELECT id FROM users WHERE username = ANY($1)`, usernames)
	if err != nil {
		return nil, err
	}
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if _, err := tx.Exec(ctx, `INSERT INTO mentions (yap_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, yapID, id); err != nil {
			return nil, err
		}
	}
	return ids, nil
}
