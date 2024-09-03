-- +goose Up
ALTER TABLE rssfeed_posts
ADD COLUMN itemcontent text;

-- +goose Down
ALTER TABLE rssfeed_posts
DROP COLUMN itemcontent;