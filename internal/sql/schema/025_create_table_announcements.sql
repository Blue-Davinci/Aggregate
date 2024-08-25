-- +goose Up
CREATE TABLE announcements (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP(0) WITH time zone DEFAULT NOW(),
    expires_at TIMESTAMP(0) WITH time zone DEFAULT NULL,
    updated_at TIMESTAMP(0) WITH time zone DEFAULT NOW(),
    created_by BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN DEFAULT TRUE,
    urgency TEXT NOT NULL CHECK (urgency IN ('low', 'medium', 'high'))
);

-- +goose Down
DROP TABLE announcements;
