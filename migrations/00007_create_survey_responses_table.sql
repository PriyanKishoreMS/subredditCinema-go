-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS survey_responses (
    id SERIAL PRIMARY KEY,
    survey_id INT NOT NULL,
    reddit_uid VARCHAR(255) NOT NULL,
    response_data JSONB NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    FOREIGN KEY (survey_id) REFERENCES surveys(id) ON DELETE CASCADE,
    FOREIGN KEY (reddit_uid) REFERENCES users(reddit_uid),
    UNIQUE (survey_id, reddit_uid)
);

CREATE INDEX IF NOT EXISTS idx_survey_responses_survey_id ON survey_responses(survey_id);
CREATE INDEX IF NOT EXISTS idx_survey_responses_reddit_uid ON survey_responses(reddit_uid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS survey_responses;
-- +goose StatementEnd
