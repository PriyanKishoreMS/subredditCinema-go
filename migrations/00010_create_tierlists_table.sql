-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tierlists (
    id SERIAL PRIMARY KEY,
    reddit_uid VARCHAR(255) NOT NULL REFERENCES users(reddit_uid),
    subreddit VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    tiers JSONB NOT NULL,
    UNIQUE (reddit_uid, title)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tierlists cascade;
-- +goose StatementEnd
