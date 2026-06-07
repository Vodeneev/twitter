-- +goose Up
-- +goose StatementBegin

CREATE TABLE hashtags (
    id          BIGSERIAL   PRIMARY KEY,
    tag         CITEXT      NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE yap_hashtags (
    yap_id      UUID        NOT NULL REFERENCES yaps(id) ON DELETE CASCADE,
    hashtag_id  BIGINT      NOT NULL REFERENCES hashtags(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (yap_id, hashtag_id)
);
CREATE INDEX yap_hashtags_hashtag_idx ON yap_hashtags(hashtag_id, created_at DESC);

CREATE TABLE mentions (
    yap_id      UUID        NOT NULL REFERENCES yaps(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (yap_id, user_id)
);
CREATE INDEX mentions_user_idx ON mentions(user_id, created_at DESC);

-- Notifications: like, follow, reply, repost, quote, mention.
CREATE TABLE notifications (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT        NOT NULL CHECK (type IN ('like','follow','reply','repost','quote','mention')),
    yap_id      UUID        REFERENCES yaps(id) ON DELETE CASCADE,
    read_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (user_id <> actor_id)
);
CREATE INDEX notifications_user_created_idx ON notifications(user_id, created_at DESC);
CREATE INDEX notifications_user_unread_idx ON notifications(user_id) WHERE read_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS mentions;
DROP TABLE IF EXISTS yap_hashtags;
DROP TABLE IF EXISTS hashtags;
-- +goose StatementEnd
