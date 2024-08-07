-- +goose Up
CREATE TABLE failed_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGSERIAL NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE, -- Optional, if you want to link to the subscription
    attempt_date TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    authorization_code VARCHAR(50),
    reference TEXT NOT NULL, -- Unique reference for the transaction attempt
    amount DECIMAL(10, 2) NOT NULL, -- The amount that was attempted to be charged
    failure_reason TEXT, -- A brief description of the failure
    error_code VARCHAR(50), -- Specific error code returned from the payment gateway
    card_last4 VARCHAR(4), -- Last 4 digits of the card used
    card_exp_month VARCHAR(2),
    card_exp_year VARCHAR(4),
    card_type VARCHAR(20),
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE failed_transactions;