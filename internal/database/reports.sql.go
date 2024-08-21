// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: reports.sql

package database

import (
	"context"
	"database/sql"
	"time"
)

const adminSubscriptionSingleReports = `-- name: AdminSubscriptionSingleReports :one
WITH
active_subscriptions AS (
    SELECT COALESCE(COUNT(*), 0) AS total_active_subscriptions
    FROM subscriptions
    WHERE status = 'active'
),

subscription_churn_rate AS (
    SELECT COALESCE(COUNT(*)::float / NULLIF(COUNT(*) OVER (), 0)::float, 0) AS churn_rate
    FROM subscriptions
    WHERE status = 'cancelled' AND end_date >= NOW() - INTERVAL '30 days'
),

arpu AS (
    SELECT COALESCE(AVG(price), 0.0) AS average_revenue_per_user
    FROM subscriptions
),

average_subscription_duration AS (
    SELECT COALESCE(AVG(end_date - start_date), INTERVAL '0 seconds') AS average_duration
    FROM subscriptions
    WHERE status = 'cancelled'
),

total_revenue AS (
    SELECT COALESCE(SUM(price), 0) AS total_revenue
    FROM subscriptions
),

most_popular_payment_method AS (
    SELECT 
        COALESCE(payment_method, '') AS payment_method
    FROM subscriptions
    GROUP BY payment_method
    ORDER BY COUNT(*) DESC
    LIMIT 1
),

most_subscribed_plan AS (
    SELECT 
        COALESCE(plan_id, 0) AS plan_id
    FROM subscriptions
    GROUP BY plan_id
    ORDER BY COUNT(*) DESC
    LIMIT 1
),

challenged_transactions_rate AS (
    SELECT COALESCE(COUNT(*)::float / NULLIF((SELECT COUNT(*) FROM subscriptions)::float, 0), 0) AS challenged_rate
    FROM challenged_transactions
),

failed_transactions AS (
    SELECT COALESCE(COUNT(*), 0) AS failed_transactions_count
    FROM subscriptions
    WHERE status = 'failed'
),

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
    (SELECT near_expiry_count FROM subscriptions_near_expiry) AS near_expiry_count
`

type AdminSubscriptionSingleReportsRow struct {
	TotalActiveSubscriptions interface{}
	ChurnRate                interface{}
	AverageRevenuePerUser    interface{}
	AverageDuration          string
	TotalRevenue             interface{}
	MostPopularPaymentMethod string
	MostSubscribedPlan       int32
	ChallengedRate           interface{}
	FailedTransactionsCount  interface{}
	NearExpiryCount          interface{}
}

// Total Active Subscriptions
// Subscription Churn Rate
// Average Revenue Per User (ARPU)
// Average Subscription Duration
// Total Revenue
// Most Popular Payment Method
// Most Subscribed Plan
// Challenged Transactions Rate
// Failed Transactions Count
// Subscriptions Near Expiry (next 30 days)
func (q *Queries) AdminSubscriptionSingleReports(ctx context.Context) (AdminSubscriptionSingleReportsRow, error) {
	row := q.db.QueryRowContext(ctx, adminSubscriptionSingleReports)
	var i AdminSubscriptionSingleReportsRow
	err := row.Scan(
		&i.TotalActiveSubscriptions,
		&i.ChurnRate,
		&i.AverageRevenuePerUser,
		&i.AverageDuration,
		&i.TotalRevenue,
		&i.MostPopularPaymentMethod,
		&i.MostSubscribedPlan,
		&i.ChallengedRate,
		&i.FailedTransactionsCount,
		&i.NearExpiryCount,
	)
	return i, err
}

const challengedTransactionsOutcome = `-- name: ChallengedTransactionsOutcome :many
SELECT 
    status, 
    COUNT(*) AS outcome_count
FROM challenged_transactions
GROUP BY status
`

type ChallengedTransactionsOutcomeRow struct {
	Status       string
	OutcomeCount int64
}

func (q *Queries) ChallengedTransactionsOutcome(ctx context.Context) ([]ChallengedTransactionsOutcomeRow, error) {
	rows, err := q.db.QueryContext(ctx, challengedTransactionsOutcome)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ChallengedTransactionsOutcomeRow
	for rows.Next() {
		var i ChallengedTransactionsOutcomeRow
		if err := rows.Scan(&i.Status, &i.OutcomeCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const newVsCancelledSubscriptions = `-- name: NewVsCancelledSubscriptions :one
SELECT 
    COALESCE((SELECT COUNT(*) FROM subscriptions WHERE start_date >= NOW() - INTERVAL '30 days'), 0) AS new_subscriptions,
    COALESCE((SELECT COUNT(*) FROM subscriptions WHERE status = 'cancelled' AND end_date >= NOW() - INTERVAL '30 days'), 0) AS cancelled_subscriptions
`

type NewVsCancelledSubscriptionsRow struct {
	NewSubscriptions       interface{}
	CancelledSubscriptions interface{}
}

func (q *Queries) NewVsCancelledSubscriptions(ctx context.Context) (NewVsCancelledSubscriptionsRow, error) {
	row := q.db.QueryRowContext(ctx, newVsCancelledSubscriptions)
	var i NewVsCancelledSubscriptionsRow
	err := row.Scan(&i.NewSubscriptions, &i.CancelledSubscriptions)
	return i, err
}

const revenueByPaymentMethod = `-- name: RevenueByPaymentMethod :many
SELECT 
    payment_method, 
    COALESCE(SUM(price), 0) AS revenue
FROM subscriptions
GROUP BY payment_method
`

type RevenueByPaymentMethodRow struct {
	PaymentMethod sql.NullString
	Revenue       interface{}
}

func (q *Queries) RevenueByPaymentMethod(ctx context.Context) ([]RevenueByPaymentMethodRow, error) {
	rows, err := q.db.QueryContext(ctx, revenueByPaymentMethod)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RevenueByPaymentMethodRow
	for rows.Next() {
		var i RevenueByPaymentMethodRow
		if err := rows.Scan(&i.PaymentMethod, &i.Revenue); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const revenueByPlan = `-- name: RevenueByPlan :many
SELECT 
    pp.name AS plan_name, 
    CAST(SUM(s.price) AS BIGINT) AS total_revenue
FROM subscriptions s
JOIN payment_plans pp ON s.plan_id = pp.id
GROUP BY pp.name
`

type RevenueByPlanRow struct {
	PlanName     string
	TotalRevenue int64
}

func (q *Queries) RevenueByPlan(ctx context.Context) ([]RevenueByPlanRow, error) {
	rows, err := q.db.QueryContext(ctx, revenueByPlan)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []RevenueByPlanRow
	for rows.Next() {
		var i RevenueByPlanRow
		if err := rows.Scan(&i.PlanName, &i.TotalRevenue); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const subscriptionsByCurrency = `-- name: SubscriptionsByCurrency :many
SELECT 
    currency, 
    COALESCE(COUNT(*), 0) AS currency_count
FROM subscriptions
GROUP BY currency
`

type SubscriptionsByCurrencyRow struct {
	Currency      sql.NullString
	CurrencyCount interface{}
}

func (q *Queries) SubscriptionsByCurrency(ctx context.Context) ([]SubscriptionsByCurrencyRow, error) {
	rows, err := q.db.QueryContext(ctx, subscriptionsByCurrency)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubscriptionsByCurrencyRow
	for rows.Next() {
		var i SubscriptionsByCurrencyRow
		if err := rows.Scan(&i.Currency, &i.CurrencyCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const subscriptionsOverTime = `-- name: SubscriptionsOverTime :many
SELECT 
    date_trunc('day', start_date)::date AS day, 
    COALESCE(COUNT(*), 0) AS subscriptions_count
FROM subscriptions
WHERE start_date >= NOW() - INTERVAL '30 days'
GROUP BY day
ORDER BY day
`

type SubscriptionsOverTimeRow struct {
	Day                time.Time
	SubscriptionsCount interface{}
}

func (q *Queries) SubscriptionsOverTime(ctx context.Context) ([]SubscriptionsOverTimeRow, error) {
	rows, err := q.db.QueryContext(ctx, subscriptionsOverTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SubscriptionsOverTimeRow
	for rows.Next() {
		var i SubscriptionsOverTimeRow
		if err := rows.Scan(&i.Day, &i.SubscriptionsCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}