-- +goose Up
CREATE TABLE payment_plans (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    features TEXT[] NOT NULL,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    comment TEXT
);
-- Add our pre-made payment plans
INSERT INTO payment_plans (name, price, features, comment)
VALUES 
    ('Free', 0.00, ARRAY['5 feeds', 'Follow 10 feeds', '10 messages/day'], 'Get Started'),
    ('Monthly', 10.00, ARRAY['20 feeds', 'Follow 40 feeds', '20 messages/day'], 'Buy Now'),
    ('Annual', 100.00, ARRAY['Unlimited feeds', 'Follow unlimited feeds', 'Unlimited messages/day'], 'Buy Now');


-- +goose Down
DROP TABLE payment_plans;