-- +goose Up
ALTER TABLE feeds ADD COLUMN img_url TEXT NOT NULL DEFAULT 'https://cdn.pixabay.com/photo/2017/06/25/14/38/rss-2440955_960_720.png';
ALTER TABLE feeds ADD COLUMN last_fetched_at timestamp(0);

-- +goose Down
ALTER TABLE feeds DROP COLUMN img_url;
ALTER TABLE feeds DROP COLUMN last_fetched_at;