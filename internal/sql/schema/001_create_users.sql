-- +goose Up
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users(
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    activated bool NOT NULL DEFAULT FALSE,
    version integer NOT NULL DEFAULT 1
);

-- +goose Down
DROP TABLE users;