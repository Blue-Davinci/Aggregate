-- +goose Up
CREATE TABLE feeds(
    id UUID PRIMARY KEY,
        created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
        updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
        name TEXT NOT NULL,
        url TEXT UNIQUE NOT NULL,
        version INT NOT NULL DEFAULT 1,
        user_id bigserial NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;