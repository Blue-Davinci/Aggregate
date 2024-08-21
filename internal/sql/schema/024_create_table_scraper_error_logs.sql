-- +goose Up
CREATE TABLE scraper_error_logs (
    id SERIAL PRIMARY KEY,
    error_type VARCHAR(255) NOT NULL,
    message TEXT,
    feed_url TEXT,
    occurred_at TIMESTAMP DEFAULT NOW(),
    status_code INTEGER,
    retry_attempts INTEGER DEFAULT 0,
    admin_notified BOOLEAN DEFAULT FALSE,
    resolved BOOLEAN DEFAULT FALSE,
    resolution_notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    occurrence_count INTEGER DEFAULT 1,
    last_occurrence TIMESTAMP DEFAULT NOW(),
    UNIQUE (error_type, feed_url)
);



-- +goose Down
DROP TABLE scraper_error_logs;