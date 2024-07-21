-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS poll_votes (
    id SERIAL PRIMARY KEY,
    poll_id INT NOT NULL,
    reddit_uid VARCHAR(255) NOT NULL,
    option_id INT NOT NULL, 
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE,
    FOREIGN KEY (reddit_uid) REFERENCES users(reddit_uid),
    UNIQUE (poll_id, reddit_uid)
);
CREATE INDEX IF NOT EXISTS idx_poll_votes_poll_id ON poll_votes(poll_id);
CREATE INDEX IF NOT EXISTS idx_poll_votes_poll_id_option_id ON poll_votes(poll_id, option_id)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS poll_votes
-- +goose StatementEnd
