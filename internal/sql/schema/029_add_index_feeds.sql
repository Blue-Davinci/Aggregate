-- +goose Up
CREATE INDEX idx_feeds_is_hidden ON feeds (is_hidden);

CREATE INDEX idx_feeds_priority_created_at ON feeds (priority, created_at);

CREATE INDEX idx_feeds_approval_status ON feeds (approval_status);

CREATE INDEX idx_feeds_feed_type ON feeds (feed_type);

-- +goose Down
DROP INDEX IF EXISTS idx_feeds_is_hidden;

DROP INDEX IF EXISTS idx_feeds_priority_created_at;

DROP INDEX IF EXISTS idx_feeds_approval_status;

DROP INDEX IF EXISTS idx_feeds_feed_type;
