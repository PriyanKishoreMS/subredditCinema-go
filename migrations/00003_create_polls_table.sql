-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS polls (
    id SERIAL PRIMARY KEY,
    reddit_uid VARCHAR(255) NOT NULL,
    subreddit VARCHAR(128) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    options JSONB NOT NULL, 
    start_time timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    end_time timestamp(0) with time zone NOT NULL,
    FOREIGN KEY (reddit_uid) REFERENCES users(reddit_uid)
);

CREATE INDEX IF NOT EXISTS idx_polls_reddit_uid ON polls(reddit_uid);
CREATE INDEX IF NOT EXISTS idx_polls_subreddit ON polls(subreddit);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS polls
-- +goose StatementEnd
