-- +goose Up
ALTER TABLE payment_plans
    ALTER COLUMN status TYPE VARCHAR(20),
    ADD CONSTRAINT status_check CHECK (status IN ('active', 'inactive')),
    ADD CONSTRAINT unique_name UNIQUE (name);

-- +goose Down
ALTER TABLE payment_plans
    DROP CONSTRAINT status_check,
    DROP CONSTRAINT unique_name,
    ALTER COLUMN status TYPE TEXT;