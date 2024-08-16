-- +goose Up
ALTER TABLE permissions
ADD CONSTRAINT unique_code UNIQUE (code);

-- +goose Down
ALTER TABLE permissions
DROP CONSTRAINT unique_code;