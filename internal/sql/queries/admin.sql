-- name: AdminGetAllUsers :many
SELECT
    u.id,
    u.created_at,
    u.name,
    u.email,
    u.password_hash,
    u.activated,
    u.version,
    u.user_img,
    COALESCE(
        NULLIF(array_agg(p.code::TEXT ORDER BY p.id), '{}'), 
        ARRAY['normal user']
    ) AS permissions,
    COUNT(*) OVER() AS total_count
FROM
    public.users u
LEFT JOIN
    users_permissions up ON up.user_id = u.id
LEFT JOIN
    permissions p ON p.id = up.permission_id
WHERE
    ($1 = '' OR to_tsvector('simple', u.name) @@ plainto_tsquery('simple', $1))
GROUP BY
    u.id, u.created_at, u.name, u.email, u.password_hash, u.activated, u.version, u.user_img
ORDER BY
    u.created_at DESC
LIMIT $2 OFFSET $3;

-- name: AdminGetAllPaymentPlans :many
SELECT id, name, image, description, duration, price, features, created_at, updated_at, status,version
FROM payment_plans;

-- name: AdminGetAllSubscriptions :many
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
    s.updated_at,
    CASE 
        WHEN EXISTS (
            SELECT 1
            FROM challenged_transactions ct
            WHERE ct.referenced_subscription_id = s.id
        ) THEN true
        ELSE false
    END AS has_challenged_transactions
FROM 
    subscriptions s
JOIN 
    payment_plans p ON s.plan_id = p.id
ORDER BY 
    s.start_date DESC
LIMIT $1 OFFSET $2;


-- name: AdminGetChallengedTransactionsBySubscriptionID :many
SELECT 
    ct.id AS transaction_id,
    ct.user_id,
    u.name AS user_name,
    u.email AS user_email,
    u.user_img AS user_img,
    ct.referenced_subscription_id,
    ct.authorization_url,
    ct.reference,
    ct.created_at,
    ct.updated_at,
    ct.status,
    s.plan_id,
    s.price,
    s.start_date,
    s.end_date
FROM 
    challenged_transactions ct
JOIN 
    subscriptions s 
ON 
    ct.referenced_subscription_id = s.id
JOIN 
    users u
ON 
    ct.user_id = u.id
WHERE 
    ct.referenced_subscription_id = $1;

-- name: AdminGetStatistics :one
WITH
-- Get statistics from the users table
user_stats AS (
    SELECT
        COUNT(*) AS total_users,
        COUNT(*) FILTER (WHERE activated = true) AS active_users,
        COUNT(*) FILTER (WHERE created_at >= now() - INTERVAL '7 days') AS new_signups
    FROM users
),
-- Get statistics from the subscriptions table
subscription_stats AS (
    SELECT
        COALESCE(SUM(price), 0)::numeric AS total_revenue,
        COUNT(*) FILTER (WHERE status = 'active') AS active_subscriptions,
        COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled_subscriptions,
        COUNT(*) FILTER (WHERE status = 'expired') AS expired_subscriptions,
        COALESCE(
            (SELECT payment_method FROM subscriptions
            GROUP BY payment_method
            ORDER BY COUNT(*) DESC
            LIMIT 1),
            'N/A'
        ) AS most_used_payment_method
    FROM subscriptions
),
-- Get statistics from the comments table
comment_stats AS (
    SELECT
        COUNT(*) AS total_comments,
        COUNT(*) FILTER (WHERE created_at >= now() - INTERVAL '7 days') AS recent_comments
    FROM comments
)
-- Combine all the statistics
SELECT
    us.total_users,
    us.active_users,
    us.new_signups,
    ss.total_revenue,
    ss.active_subscriptions,
    ss.cancelled_subscriptions,
    ss.expired_subscriptions,
    ss.most_used_payment_method,
    cs.total_comments,
    cs.recent_comments
FROM user_stats us, subscription_stats ss, comment_stats cs;


-- name: AdminCreatePaymentPlan :one
INSERT INTO payment_plans (
    name, image, description, duration, price, features, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: AdminUpdatePaymentPlan :one
UPDATE payment_plans
SET 
    name = $1,
    image = $2,
    description = $3,
    duration = $4,
    price = $5,
    features = $6,
    status = $7,
    version = version + 1,
    updated_at = now()
WHERE 
    id = $8 AND version = $9
RETURNING version;

-- name: AdminGetPaymentPlanByID :one
SELECT id, name, image, description, duration, price, features, created_at, updated_at, status, version
FROM payment_plans
WHERE id = $1;

-- name: AdminCreateNewPermission :one
INSERT INTO permissions (code)
VALUES ($1)
RETURNING id, code;

-- name: AdminUpdatePermissionCode :one
UPDATE permissions
SET code = $2
WHERE id = $1
RETURNING id, code;

-- name: AdminDeletePermission :one
DELETE FROM permissions
WHERE id = $1
RETURNING id, code;

