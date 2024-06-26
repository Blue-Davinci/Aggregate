-- +goose Up

CREATE TABLE IF NOT EXISTS api_keys(
    api_key bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL DEFAULT NOW() + INTERVAL '3 day',
    scope text NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS api_keys;