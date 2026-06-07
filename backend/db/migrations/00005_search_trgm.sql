-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS yaps_content_trgm_idx ON yaps USING GIN (lower(content) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_username_trgm_idx ON users USING GIN (lower(username) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users_display_name_trgm_idx ON users USING GIN (lower(display_name) gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users_display_name_trgm_idx;
DROP INDEX IF EXISTS users_username_trgm_idx;
DROP INDEX IF EXISTS yaps_content_trgm_idx;
-- +goose StatementEnd
