-- +goose Up
CREATE TABLE feed_rejections (
    id BIGSERIAL PRIMARY KEY,
    feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE,
    rejected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    rejected_by BIGSERIAL REFERENCES users(id) ON DELETE SET NULL, -- user who rejected the feed
    reason TEXT NOT NULL
);

CREATE INDEX idx_feed_rejections_feed_id ON feed_rejections(feed_id);
CREATE INDEX idx_feed_rejections_rejected_by ON feed_rejections(rejected_by);

-- +goose Down
DROP TABLE feed_rejections;
