-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reddit_posts (
    id VARCHAR(32) PRIMARY KEY,
    post_id BIGSERIAL UNIQUE NOT NULL,
    name VARCHAR(32) UNIQUE NOT NULL,
    created_utc TIMESTAMP NOT NULL,
    permalink VARCHAR(255) NOT NULL,
    title TEXT NOT NULL,
    category VARCHAR(32) NOT NULL,
    top_and_controversial BOOLEAN DEFAULT FALSE,
    selftext TEXT NOT NULL,
    score INT NOT NULL,
    upvote_ratio FLOAT NOT NULL,
    num_comments INT NOT NULL,
    subreddit VARCHAR(32) NOT NULL,
    subreddit_id VARCHAR(32) NOT NULL,
    subreddit_subscribers BIGINT NOT NULL,
    author VARCHAR(64) NOT NULL,
    author_fullname VARCHAR(32) NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_reddit_posts_post_id ON reddit_posts(post_id);
CREATE INDEX IF NOT EXISTS idx_reddit_posts_title ON reddit_posts(title);
CREATE INDEX IF NOT EXISTS idx_reddit_posts_subreddit ON reddit_posts(subreddit);
CREATE INDEX IF NOT EXISTS idx_reddit_posts_category ON reddit_posts(category);
CREATE INDEX IF NOT EXISTS idx_reddit_posts_created_utc ON reddit_posts(created_utc) -- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reddit_posts;
-- +goose StatementEnd