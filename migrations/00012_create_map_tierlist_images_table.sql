-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tierlist_images_map (
    id SERIAL PRIMARY KEY,
    tierlist_id INT NOT NULL REFERENCES tierlists(id) ON DELETE CASCADE,
    image_id INT NOT NULL REFERENCES tierlist_images(id) ON DELETE CASCADE,
    UNIQUE(tierlist_id, image_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tierlist_images_map;
-- +goose StatementEnd
