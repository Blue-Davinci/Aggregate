-- name: AdminSubscriptionSingleReports :one
WITH
-- Total Active Subscriptions
active_subscriptions AS (
    SELECT COALESCE(COUNT(*), 0) AS total_active_subscriptions
    FROM subscriptions
    WHERE status = 'active'
),

-- Subscription Churn Rate
subscription_churn_rate AS (
    SELECT COALESCE(COUNT(*)::float / NULLIF(COUNT(*) OVER (), 0)::float, 0) AS churn_rate
    FROM subscriptions
    WHERE status = 'cancelled' AND end_date >= NOW() - INTERVAL '30 days'
),

-- Average Revenue Per User (ARPU)
arpu AS (
    SELECT COALESCE(AVG(price), 0.0) AS average_revenue_per_user
    FROM subscriptions
),

-- Average Subscription Duration
average_subscription_duration AS (
    SELECT COALESCE(AVG(end_date - start_date), INTERVAL '0 seconds') AS average_duration
    FROM subscriptions
    WHERE status = 'cancelled'
),

-- Total Revenue
total_revenue AS (
    SELECT COALESCE(SUM(price), 0) AS total_revenue
    FROM subscriptions
),

-- Most Popular Payment Method
most_popular_payment_method AS (
    SELECT 
        COALESCE(payment_method, '') AS payment_method
    FROM subscriptions
    GROUP BY payment_method
    ORDER BY COUNT(*) DESC
    LIMIT 1
),

-- Most Subscribed Plan
most_subscribed_plan AS (
    SELECT 
        COALESCE(plan_id, 0) AS plan_id
    FROM subscriptions
    GROUP BY plan_id
    ORDER BY COUNT(*) DESC
    LIMIT 1
),

-- Challenged Transactions Rate
challenged_transactions_rate AS (
    SELECT COALESCE(COUNT(*)::float / NULLIF((SELECT COUNT(*) FROM subscriptions)::float, 0), 0) AS challenged_rate
    FROM challenged_transactions
),

-- Failed Transactions Count
failed_transactions AS (
    SELECT COALESCE(COUNT(*), 0) AS failed_transactions_count
    FROM subscriptions
    WHERE status = 'failed'
),

-- Subscriptions Near Expiry (next 30 days)
subscriptions_near_expiry AS (
    SELECT COALESCE(COUNT(*), 0) AS near_expiry_count
    FROM subscriptions
    WHERE end_date BETWEEN NOW() AND NOW() + INTERVAL '30 days'
)

SELECT 
    (SELECT total_active_subscriptions FROM active_subscriptions) AS total_active_subscriptions,
    (SELECT churn_rate FROM subscription_churn_rate) AS churn_rate,
    (SELECT average_revenue_per_user FROM arpu) AS average_revenue_per_user,
    (SELECT EXTRACT(EPOCH FROM average_duration) FROM average_subscription_duration) AS average_duration,
    (SELECT total_revenue FROM total_revenue) AS total_revenue,
    (SELECT payment_method FROM most_popular_payment_method) AS most_popular_payment_method,
    (SELECT plan_id FROM most_subscribed_plan) AS most_subscribed_plan,
    (SELECT challenged_rate FROM challenged_transactions_rate) AS challenged_rate,
    (SELECT failed_transactions_count FROM failed_transactions) AS failed_transactions_count,
    (SELECT near_expiry_count FROM subscriptions_near_expiry) AS near_expiry_count;

-- name: RevenueByPlan :many
SELECT 
    pp.name AS plan_name, 
    CAST(SUM(s.price) AS BIGINT) AS total_revenue
FROM subscriptions s
JOIN payment_plans pp ON s.plan_id = pp.id
GROUP BY pp.name;


-- name: ChallengedTransactionsOutcome :many
SELECT 
    status, 
    COUNT(*) AS outcome_count
FROM challenged_transactions
GROUP BY status;

-- name: RevenueByPaymentMethod :many
SELECT 
    payment_method, 
    COALESCE(SUM(price), 0) AS revenue
FROM subscriptions
GROUP BY payment_method;

-- name: SubscriptionsOverTime :many
SELECT 
    date_trunc('day', start_date)::date AS day, 
    COALESCE(COUNT(*), 0) AS subscriptions_count
FROM subscriptions
WHERE start_date >= NOW() - INTERVAL '30 days'
GROUP BY day
ORDER BY day;

-- name: NewVsCancelledSubscriptions :one
SELECT 
    COALESCE((SELECT COUNT(*) FROM subscriptions WHERE start_date >= NOW() - INTERVAL '30 days'), 0) AS new_subscriptions,
    COALESCE((SELECT COUNT(*) FROM subscriptions WHERE status = 'cancelled' AND end_date >= NOW() - INTERVAL '30 days'), 0) AS cancelled_subscriptions;

-- name: SubscriptionsByCurrency :many
SELECT 
    currency, 
    COALESCE(COUNT(*), 0) AS currency_count
FROM subscriptions
GROUP BY currency;
