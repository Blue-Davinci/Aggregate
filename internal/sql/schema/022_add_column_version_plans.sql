-- +goose Up
ALTER TABLE payment_plans
ADD COLUMN version integer NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE payment_plans
DROP COLUMN version;