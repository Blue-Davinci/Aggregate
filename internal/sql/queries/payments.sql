-- name: CreateSubscription :one
INSERT INTO subscriptions (
		user_id, plan_id, start_date, end_date, price, status, 
		transaction_id, payment_method, authorization_code, 
		card_last4, card_exp_month, card_exp_year, card_type, currency
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
RETURNING id, created_at, updated_at;

-- name: GetPaymentPlans :many
SELECT id, name, image, description, duration, price, features, created_at, updated_at, status
FROM payment_plans
WHERE status = 'active';

-- name: GetPaymentPlanByID :one
SELECT id, name, image, description, duration, price, features, created_at, updated_at, status
FROM payment_plans
WHERE id = $1 AND status = 'active';

-- name: GetSubscriptionByID :one
SELECT id, user_id, plan_id, start_date, end_date, price, status
FROM subscriptions
WHERE user_id = $1 AND status = 'active' AND end_date > NOW();

-- name: GetAllSubscriptionsByID :many
SELECT 
    count(*) OVER() as total_records,
    s.id, 
    s.user_id, 
    s.plan_id, 
    p.name as plan_name,
    p.image as plan_image,
    p.duration as plan_duration,
    s.start_date, 
    s.end_date, 
    s.price, 
    s.status, 
    s.transaction_id, 
    s.payment_method, 
    s.card_last4, 
    s.card_exp_month, 
    s.card_exp_year, 
    s.card_type, 
    s.currency, 
    s.created_at, 
    s.updated_at
FROM 
    subscriptions s
JOIN 
    payment_plans p ON s.plan_id = p.id
WHERE 
    s.user_id = $1 
ORDER BY 
    s.start_date DESC
LIMIT $2 OFFSET $3;

