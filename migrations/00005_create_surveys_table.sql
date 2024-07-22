-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS surveys (
    id SERIAL PRIMARY KEY,
    reddit_uid VARCHAR(255) NOT NULL,
    subreddit VARCHAR(128) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    end_time timestamp(0) with time zone,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    FOREIGN KEY (reddit_uid) REFERENCES users(reddit_uid)
);

CREATE INDEX IF NOT EXISTS idx_surveys_reddit_uid ON surveys(reddit_uid);
CREATE INDEX IF NOT EXISTS idx_surveys_subreddit ON surveys(subreddit);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS surveys;
-- +goose StatementEnd
