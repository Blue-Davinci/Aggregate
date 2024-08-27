-- +goose Up
ALTER TABLE feeds
ADD COLUMN approval_status TEXT NOT NULL DEFAULT 'pending',
ADD COLUMN priority TEXT NOT NULL DEFAULT 'normal';

-- Add constraints for approved column
ALTER TABLE feeds
ADD CONSTRAINT chk_approved CHECK (approval_status IN ('approved', 'pending', 'rejected'));

-- Add constraints for priority column
ALTER TABLE feeds
ADD CONSTRAINT chk_priority CHECK (priority IN ('high', 'normal'));


-- +goose Down
ALTER TABLE feeds
DROP COLUMN approval_status,
DROP COLUMN priority;
