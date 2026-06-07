-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

-- Users: anyone who registered. The handle is `username`.
CREATE TABLE users (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email               CITEXT      NOT NULL UNIQUE,
    username            CITEXT      NOT NULL UNIQUE,
    display_name        TEXT        NOT NULL DEFAULT '',
    password_hash       TEXT        NOT NULL,
    bio                 TEXT        NOT NULL DEFAULT '',
    location            TEXT        NOT NULL DEFAULT '',
    website             TEXT        NOT NULL DEFAULT '',
    avatar_key          TEXT,
    header_key          TEXT,
    is_admin            BOOLEAN     NOT NULL DEFAULT FALSE,
    is_banned           BOOLEAN     NOT NULL DEFAULT FALSE,
    followers_count     INT         NOT NULL DEFAULT 0,
    following_count     INT         NOT NULL DEFAULT 0,
    yaps_count          INT         NOT NULL DEFAULT 0,
    email_verified_at   TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX users_username_idx ON users(username);

-- Server-side sessions. The opaque session id is stored in an HttpOnly cookie.
CREATE TABLE sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at      TIMESTAMPTZ NOT NULL,
    remember_me     BOOLEAN     NOT NULL DEFAULT FALSE,
    user_agent      TEXT,
    ip              INET,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);

-- Email verification: hashed, single-use, time-bound tokens.
CREATE TABLE email_verifications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX email_verifications_user_id_idx ON email_verifications(user_id);

-- Password reset: hashed, single-use, time-bound tokens.
CREATE TABLE password_resets (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  TEXT        NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX password_resets_user_id_idx ON password_resets(user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
