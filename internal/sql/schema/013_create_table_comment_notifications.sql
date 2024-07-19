-- +goose Up
CREATE TABLE comment_notifications (
    id SERIAL PRIMARY KEY,
    comment_id UUID NOT NULL,
    post_id UUID NOT NULL,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT now(),
    FOREIGN KEY (post_id) REFERENCES rssfeed_posts(id) ON DELETE CASCADE,
    FOREIGN KEY (comment_id) REFERENCES comments(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE comment_notifications;