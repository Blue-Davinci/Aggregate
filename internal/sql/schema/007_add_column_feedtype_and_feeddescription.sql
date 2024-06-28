-- +goose Up
ALTER TABLE feeds ADD COLUMN feed_type TEXT NOT NULL DEFAULT 'general';
ALTER TABLE feeds ADD COLUMN feed_description TEXT NOT NULL DEFAULT 'An interesting take on modern and latest changes in this field';

-- +goose Down
ALTER TABLE feeds DROP COLUMN feed_type;
ALTER TABLE feeds DROP COLUMN feed_description;