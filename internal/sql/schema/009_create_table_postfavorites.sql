-- +goose Up
CREATE TABLE postfavorites (
    id BIGSERIAL PRIMARY KEY,
    post_id UUID UNIQUE NOT NULL,
    feed_id UUID NOT NULL,
    user_id BIGINT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT now(),
    FOREIGN KEY (post_id) REFERENCES rssfeed_posts(id) ON DELETE CASCADE,
    FOREIGN KEY (feed_id) REFERENCES feeds(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);


-- +goose Down
DROP TABLE postfavorites;