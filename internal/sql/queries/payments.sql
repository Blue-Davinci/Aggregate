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