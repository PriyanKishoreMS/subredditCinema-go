-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tierlist_images (
    id SERIAL PRIMARY KEY,
    url VARCHAR(2048) UNIQUE NOT NULL
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tierlist_images cascade;
-- +goose StatementEnd
