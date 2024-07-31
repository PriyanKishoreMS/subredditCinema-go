-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS survey_options (
    id SERIAL PRIMARY KEY,
    question_id INT NOT NULL,
    option_order INT NOT NULL,
    option_text TEXT NOT NULL,
    FOREIGN KEY (question_id) REFERENCES survey_questions(id) ON DELETE CASCADE,
    UNIQUE (question_id, option_order)
);
CREATE INDEX IF NOT EXISTS idx_survey_options_question_id ON survey_options(question_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS survey_options;
-- +goose StatementEnd
