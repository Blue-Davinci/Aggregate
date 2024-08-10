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

-- name: GetAllExpiredSubscriptions :many
SELECT count(*) OVER() as total_records,
    s.id AS subscription_id,
    s.authorization_code,
    s.plan_id,
    s.price,
    s.currency,
    s.user_id,
    u.email,
    u.name
FROM 
    subscriptions s
JOIN 
    users u ON s.user_id = u.id
WHERE 
    s.status = 'expired' AND s.end_date < NOW()
ORDER BY 
    s.end_date ASC
LIMIT $1 OFFSET $2;

-- name: GetPendingChallengedTransactionsByUser :many
SELECT 
    id,
    user_id,
    referenced_subscription_id,
    authorization_url,
    reference,
    created_at,
    updated_at,
    status
FROM 
    challenged_transactions
WHERE 
    user_id = $1
    AND status = 'pending';

-- name: GetPendingChallengedTransactionBySubscriptionID :one
SELECT 
    id, 
    user_id, 
    referenced_subscription_id, 
    authorization_url, 
    reference, 
    created_at, 
    updated_at, 
    status
FROM 
    challenged_transactions
WHERE 
    referenced_subscription_id = $1
    AND status = 'pending';


-- name: CreateChallangedTransaction :one
INSERT INTO challenged_transactions (
    user_id,
    referenced_subscription_id,
    authorization_url,
    reference,
    status
) VALUES ($1,$2,$3, $4, $5) 
RETURNING id, created_at;

-- name: CreateFailedTransaction :one
INSERT INTO failed_transactions (
    user_id, 
    subscription_id, 
    attempt_date, 
    authorization_code, 
    reference, 
    amount, 
    failure_reason, 
    error_code, 
    card_last4, 
    card_exp_month, 
    card_exp_year, 
    card_type
) VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, created_at, updated_at;

-- name: UpdateSubscriptionStatus :one
UPDATE subscriptions
SET status = $1
WHERE id = $2
AND user_id = $3
RETURNING updated_at;

-- name: UpdateSubscriptionStatusAfterExpiration :many
UPDATE subscriptions
SET status = 'expired'
WHERE end_date < CURRENT_DATE
AND status NOT IN ('expired', 'renewed', 'cancelled')
RETURNING id, user_id, updated_at;

-- name: UpdateChallengedTransactionStatus :one
UPDATE challenged_transactions
SET status = $1, updated_at = NOW()
WHERE id = $2
AND user_id = $3
RETURNING *;