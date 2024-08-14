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
