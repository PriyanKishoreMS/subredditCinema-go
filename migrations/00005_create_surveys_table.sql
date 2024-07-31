-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS surveys (
    id SERIAL PRIMARY KEY,
    reddit_uid VARCHAR(255) NOT NULL,
    subreddit VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    end_time timestamp(0) with time zone,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_result_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    UNIQUE (reddit_uid, title),
    FOREIGN KEY (reddit_uid) REFERENCES users(reddit_uid)
);
CREATE INDEX IF NOT EXISTS idx_surveys_reddit_uid ON surveys(reddit_uid);
CREATE INDEX IF NOT EXISTS idx_surveys_subreddit ON surveys(subreddit);
CREATE INDEX IF NOT EXISTS idx_surveys_is_active ON surveys(is_active);
CREATE INDEX IF NOT EXISTS idx_surveys_start_time ON surveys(start_time);
CREATE INDEX IF NOT EXISTS idx_surveys_end_time ON surveys(end_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS surveys;
-- +goose StatementEnd
