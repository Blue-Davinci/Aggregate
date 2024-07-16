-- +goose Up
ALTER TABLE feeds ADD COLUMN is_hidden BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE feeds DROP COLUMN is_hidden;