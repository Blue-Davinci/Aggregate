-- +goose Up
CREATE TABLE challenged_transactions (
    id bigserial PRIMARY KEY,
    user_id bigserial NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    referenced_subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    authorization_url TEXT NOT NULL,
    reference TEXT NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'successful', 'failed'))
);

-- +goose Down
DROP TABLE challenged_transactions;