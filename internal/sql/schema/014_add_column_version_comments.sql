-- +goose Up
ALTER TABLE comments ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE comments DROP COLUMN version;