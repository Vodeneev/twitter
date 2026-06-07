-- +goose Up
-- +goose StatementBegin

-- Yaps: the core post entity (a "tweet").
-- reply_to_id: parent yap when this is a reply.
-- quote_of_id: quoted yap when this is a quote-yap.
CREATE TABLE yaps (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content         TEXT        NOT NULL DEFAULT '',
    reply_to_id     UUID        REFERENCES yaps(id) ON DELETE CASCADE,
    quote_of_id     UUID        REFERENCES yaps(id) ON DELETE SET NULL,
    likes_count     INT         NOT NULL DEFAULT 0,
    reposts_count   INT         NOT NULL DEFAULT 0,
    replies_count   INT         NOT NULL DEFAULT 0,
    quotes_count    INT         NOT NULL DEFAULT 0,
    bookmarks_count INT         NOT NULL DEFAULT 0,
    search_vector   TSVECTOR,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX yaps_author_created_idx ON yaps(author_id, created_at DESC);
CREATE INDEX yaps_reply_to_idx ON yaps(reply_to_id, created_at);
CREATE INDEX yaps_created_idx ON yaps(created_at DESC);
CREATE INDEX yaps_search_idx ON yaps USING GIN(search_vector);

-- Images attached to a yap (S3-compatible storage keys).
CREATE TABLE yap_media (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    yap_id      UUID        NOT NULL REFERENCES yaps(id) ON DELETE CASCADE,
    s3_key      TEXT        NOT NULL,
    position    INT         NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX yap_media_yap_id_idx ON yap_media(yap_id, position);

-- Follows: directed edge follower -> followee.
CREATE TABLE follows (
    follower_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id  UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id),
    CHECK (follower_id <> followee_id)
);
CREATE INDEX follows_followee_idx ON follows(followee_id);

-- Likes.
CREATE TABLE likes (
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    yap_id      UUID        NOT NULL REFERENCES yaps(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, yap_id)
);
CREATE INDEX likes_yap_idx ON likes(yap_id);

-- Reposts (pure retweets, no extra text). Quote-yaps live on the yaps table.
CREATE TABLE reposts (
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    yap_id      UUID        NOT NULL REFERENCES yaps(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, yap_id)
);
CREATE INDEX reposts_yap_idx ON reposts(yap_id);
CREATE INDEX reposts_user_created_idx ON reposts(user_id, created_at DESC);

-- Bookmarks (private saves).
CREATE TABLE bookmarks (
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    yap_id      UUID        NOT NULL REFERENCES yaps(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, yap_id)
);
CREATE INDEX bookmarks_user_created_idx ON bookmarks(user_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS reposts;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS follows;
DROP TABLE IF EXISTS yap_media;
DROP TABLE IF EXISTS yaps;
-- +goose StatementEnd
