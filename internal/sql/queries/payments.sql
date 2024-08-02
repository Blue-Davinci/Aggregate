-- name: CreateSubscription :one
INSERT INTO subscriptions (
		user_id, plan_id, start_date, end_date, price, status, 
		transaction_id, payment_method, authorization_code, 
		card_last4, card_exp_month, card_exp_year, card_type
) VALUES (
	$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING id, user_id, plan_id, start_date, end_date, status, transaction_id;

-- name: GetPaymentPlans :many
SELECT id, name, description, price, features, created_at, updated_at, status
FROM payment_plans
WHERE status = 'active';