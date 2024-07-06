-- +goose Up
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    feed_id UUID NOT NULL,
    feed_name TEXT NOT NULL,
    post_count INT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE notifications;