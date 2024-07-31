-- +goose Up
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    plan_id INT REFERENCES payment_plans(id) ON DELETE SET NULL,
    start_date timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    end_date timestamp(0) with time zone,
    price DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'inactive', 'canceled', 'expired')),
    transaction_id VARCHAR(50),
    payment_method VARCHAR(50), 
    authorization_code VARCHAR(50), -- Additional fields
    card_last4 VARCHAR(4),
    card_exp_month VARCHAR(2),
    card_exp_year VARCHAR(4),
    card_type VARCHAR(20),
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, plan_id)
);

CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_plan_id ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);

-- +goose Down
DROP TABLE subscriptions;
