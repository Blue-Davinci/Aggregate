-- +goose Up
-- Migration to modify the priority constraint to include 'low' in addition to 'high' and 'normal'
ALTER TABLE feeds
DROP CONSTRAINT chk_priority;

ALTER TABLE feeds
ADD CONSTRAINT chk_priority CHECK (priority IN ('high', 'normal', 'low'));

-- +goose Down
-- Revert the priority constraint back to only include 'high' and 'normal'
ALTER TABLE feeds
DROP CONSTRAINT chk_priority;

ALTER TABLE feeds
ADD CONSTRAINT chk_priority CHECK (priority IN ('high', 'normal'));
