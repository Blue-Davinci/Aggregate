-- +goose Up
CREATE TABLE announcement_reads (
    id SERIAL PRIMARY KEY,
    user_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,   
    announcement_id SERIAL REFERENCES announcements(id) ON DELETE CASCADE,
    read_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE announcement_reads;
