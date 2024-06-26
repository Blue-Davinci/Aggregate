-- +goose Up
CREATE TABLE rssfeed_posts (
    id UUID PRIMARY KEY,
    created_at timestamp(0) NOT NULL DEFAULT now(),
    updated_at timestamp(0) NOT NULL DEFAULT now(),
    channeltitle TEXT NOT NULL,
    channelurl TEXT,
    channeldescription TEXT,
    channellanguage TEXT DEFAULT 'en',
    itemtitle TEXT NOT NULL,
    itemdescription TEXT,
    itempublished_at timestamp(0) with time zone NOT NULL,
    itemurl TEXT NOT NULL UNIQUE,
    img_url TEXT NOT NULL,
    feed_id UUID NOT NULL REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE rssfeed_posts;