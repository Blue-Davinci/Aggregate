-- +goose Up

-- Create index on feed_id in rssfeed_posts
CREATE INDEX IF NOT EXISTS idx_rssfeed_posts_feed_id ON rssfeed_posts (feed_id);

-- Create index on user_id in postfavorites
CREATE INDEX IF NOT EXISTS idx_postfavorites_user_id ON postfavorites (user_id);

-- Create composite index on user_id and feed_id in feed_follows
CREATE INDEX IF NOT EXISTS idx_feed_follows_user_feed ON feed_follows (user_id, feed_id);

-- Create index on itempublished_at in rssfeed_posts
CREATE INDEX IF NOT EXISTS idx_rssfeed_posts_itempublished_at ON rssfeed_posts (itempublished_at);

-- Create index on created_at in rssfeed_posts
CREATE INDEX IF NOT EXISTS idx_rssfeed_posts_created_at ON rssfeed_posts (created_at);

-- Create index on feed_id in feed_follows
CREATE INDEX IF NOT EXISTS idx_feed_follows_feed_id ON feed_follows (feed_id);

-- Create index on url in feeds
CREATE INDEX IF NOT EXISTS idx_feeds_url ON feeds (url);

-- Create index on user_id in feeds
CREATE INDEX IF NOT EXISTS idx_feeds_user_id ON feeds (user_id);

-- Create index on name in feeds
CREATE INDEX IF NOT EXISTS idx_feeds_name ON feeds USING gin (to_tsvector('simple', name));

-- Create index on created_at in feeds
CREATE INDEX IF NOT EXISTS idx_feeds_created_at ON feeds (created_at);

-- Create index on last_fetched_at in feeds
CREATE INDEX IF NOT EXISTS idx_feeds_last_fetched_at ON feeds (last_fetched_at);


-- +goose Down

-- Drop index on feed_id in rssfeed_posts
DROP INDEX IF EXISTS idx_rssfeed_posts_feed_id;

-- Drop index on user_id in postfavorites
DROP INDEX IF EXISTS idx_postfavorites_user_id;

-- Drop composite index on user_id and feed_id in feed_follows
DROP INDEX IF EXISTS idx_feed_follows_user_feed;

-- Drop index on itempublished_at in rssfeed_posts
DROP INDEX IF EXISTS idx_rssfeed_posts_itempublished_at;

-- Drop index on created_at in rssfeed_posts
DROP INDEX IF EXISTS idx_rssfeed_posts_created_at;

-- Drop index on feed_id in feed_follows
DROP INDEX IF EXISTS idx_feed_follows_feed_id;

-- Drop index on url in feeds
DROP INDEX IF EXISTS idx_feeds_url;

-- Drop index on user_id in feeds
DROP INDEX IF EXISTS idx_feeds_user_id;

-- Drop index on name in feeds
DROP INDEX IF EXISTS idx_feeds_name;

-- Drop index on created_at in feeds
DROP INDEX IF EXISTS idx_feeds_created_at;

-- Drop index on last_fetched_at in feeds
DROP INDEX IF EXISTS idx_feeds_last_fetched_at;
