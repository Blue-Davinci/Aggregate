package data

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// RevenueByPlan represents the revenue generated by each subscription plan.
type RevenueByPlan struct {
	PlanName     string
	TotalRevenue int64
}

// ChallengedTransactionsOutcome represents the outcome of challenged transactions and the count of each status.
type ChallengedTransactionsOutcome struct {
	Status       string
	OutcomeCount int64
}

// RevenueByPaymentMethod represents the total revenue generated by each payment method.
type RevenueByPaymentMethod struct {
	PaymentMethod string
	Revenue       float64
}

// SubscriptionOverTime represents the count of subscriptions over time, grouped by day.
type SubscriptionOverTime struct {
	Day                time.Time
	SubscriptionsCount int64
}

// NewVsCancelledSubscriptions represents the count of new and canceled subscriptions within a specific period.
type NewVsCancelledSubscriptions struct {
	NewSubscriptions       int64
	CancelledSubscriptions int64
}

// SubscriptionsByCurrency represents the count of subscriptions grouped by currency.
type SubscriptionsByCurrency struct {
	Currency      string
	CurrencyCount int64
}

// SubscriptionStats contains all the reports, including single subscription statistics and multiple grouped reports.
type SubscriptionStats struct {
	SingleSubscriptionReport SingleSubscriptionReport `json:"single_subscription_report"`
	MultiReport              MultiReport              `json:"multi_report"`
}

// NewVsCancelledSubscriptionsReport encapsulates the counts of new and canceled subscriptions in the report.
type NewVsCancelledSubscriptionsReport struct {
	NewSubscriptions       int64 `json:"new_subscriptions"`
	CancelledSubscriptions int64 `json:"cancelled_subscriptions"`
}

// MultiReport represents various reports with multiple data points, such as revenue by plan, challenged transactions, etc.
type MultiReport struct {
	RevenueByPlan           []RevenueByPlan                 `json:"revenue_by_plan"`
	ChallengedTransactions  []ChallengedTransactionsOutcome `json:"challenged_transactions"`
	RevenueByPaymentMethod  []RevenueByPaymentMethod        `json:"revenue_by_payment_method"`
	SubscriptionsOverTime   []SubscriptionOverTime          `json:"subscriptions_over_time"`
	SubscriptionsByCurrency []SubscriptionsByCurrency       `json:"subscriptions_by_currency"`
}

// SingleSubscriptionReport encapsulates various single-point statistics such as total active subscriptions, churn rate, etc.
type SingleSubscriptionReport struct {
	TotalActiveSubscriptions          int64                       `json:"total_active_subscriptions"`
	ChurnRate                         float64                     `json:"churn_rate"`
	AverageRevenuePerUser             float64                     `json:"average_revenue_per_user"`
	AverageDuration                   string                      `json:"average_duration"`
	TotalRevenue                      float64                     `json:"total_revenue"`
	MostPopularPaymentMethod          string                      `json:"most_popular_payment_method"`
	MostSubscribedPlan                int64                       `json:"most_subscribed_plan"`
	ChallengedRate                    float64                     `json:"challenged_rate"`
	FailedTransactionsCount           int64                       `json:"failed_transactions_count"`
	NearExpiryCount                   int64                       `json:"near_expiry_count"`
	NewVsCancelledSubscriptionsReport NewVsCancelledSubscriptions `json:"new_vs_cancelled_subscriptions"`
}

// getSingleReports() retrieves an aggregated statistical report on the subscription data from the database.
// It returns data such as total active subscriptions, churn rate, ARPU, and other single-point metrics.
func (m AdminModel) getSingleReports(ctx context.Context) (*SingleSubscriptionReport, error) {
	// Fetch the single reports from the database.
	singleReports, err := m.DB.AdminSubscriptionSingleReports(ctx)
	if err != nil {
		return nil, err
	}

	// Convert fields to appropriate types safely.
	// Churn rate
	churnRate, err := convertToFloat64(singleReports.ChurnRate)
	if err != nil {
		return nil, err
	}
	averageRevenuePerUser, err := convertToFloat64(singleReports.AverageRevenuePerUser)
	if err != nil {
		return nil, err
	}
	totalRevenue, err := convertToFloat64(singleReports.TotalRevenue)
	if err != nil {
		return nil, err
	}
	challengedRate, err := convertToFloat64(singleReports.ChallengedRate)
	if err != nil {
		return nil, err
	}
	failedTransactionsCount, err := convertToInt64(singleReports.FailedTransactionsCount)
	if err != nil {
		return nil, err
	}
	nearExpiryCount, err := convertToInt64(singleReports.NearExpiryCount)
	if err != nil {
		return nil, err
	}
	totalActiveSubscriptions, err := convertToInt64(singleReports.TotalActiveSubscriptions)
	if err != nil {
		return nil, err
	}

	// Get new and cancelled subscriptions data.
	newVsCancelledSubscriptions, err := m.getNewVsCancelledReport(ctx)
	if err != nil {
		return nil, err
	}

	// Build the final single subscription report.
	singleReport := &SingleSubscriptionReport{
		TotalActiveSubscriptions:          totalActiveSubscriptions,
		ChurnRate:                         churnRate,
		AverageRevenuePerUser:             averageRevenuePerUser,
		AverageDuration:                   singleReports.AverageDuration,
		TotalRevenue:                      totalRevenue,
		MostPopularPaymentMethod:          singleReports.MostPopularPaymentMethod,
		MostSubscribedPlan:                int64(singleReports.MostSubscribedPlan),
		ChallengedRate:                    challengedRate,
		FailedTransactionsCount:           failedTransactionsCount,
		NearExpiryCount:                   nearExpiryCount,
		NewVsCancelledSubscriptionsReport: *newVsCancelledSubscriptions,
	}

	return singleReport, nil
}

// getNewVsCancelledReport retrieves the count of new and canceled subscriptions from the database.
func (m AdminModel) getNewVsCancelledReport(ctx context.Context) (*NewVsCancelledSubscriptions, error) {
	// Fetch the new and cancelled subscriptions data.
	newVsCancelledSubscriptions, err := m.DB.NewVsCancelledSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	// Convert fields to appropriate types safely.
	NewSubscriptions, err := convertToInt64(newVsCancelledSubscriptions.NewSubscriptions)
	if err != nil {
		return nil, err
	}
	CancelledSubscriptions, err := convertToInt64(newVsCancelledSubscriptions.CancelledSubscriptions)
	if err != nil {
		return nil, err
	}

	// Return the structured data.
	return &NewVsCancelledSubscriptions{
		NewSubscriptions:       NewSubscriptions,
		CancelledSubscriptions: CancelledSubscriptions,
	}, nil
}

// getMultiReport retrieves various grouped subscription statistics from the database, including revenue by plan,
// challenged transactions, revenue by payment method, and more.
func (m AdminModel) getMultiReport(ctx context.Context) (*MultiReport, error) {
	// Get revenue by plan data.
	var revenueByPlanData []RevenueByPlan
	revenueByPlan, err := m.DB.RevenueByPlan(ctx)
	if err != nil {
		return nil, err
	}
	for _, plan := range revenueByPlan {
		convertedTotalRevenue, err := convertToInt64(plan.TotalRevenue)
		if err != nil {
			return nil, err
		}
		revenueByPlanData = append(revenueByPlanData, RevenueByPlan{
			PlanName:     plan.PlanName,
			TotalRevenue: convertedTotalRevenue,
		})
	}

	// Get challenged transactions outcome data.
	challengedTransactions, err := m.DB.ChallengedTransactionsOutcome(ctx)
	if err != nil {
		return nil, err
	}
	var challengedTransactionsData []ChallengedTransactionsOutcome
	for _, transaction := range challengedTransactions {
		convertedOutcomeCount, err := convertToInt64(transaction.OutcomeCount)
		if err != nil {
			return nil, err
		}
		challengedTransactionsData = append(challengedTransactionsData, ChallengedTransactionsOutcome{
			Status:       transaction.Status,
			OutcomeCount: convertedOutcomeCount,
		})
	}

	// Get revenue by payment method data.
	revenueByPaymentMethod, err := m.DB.RevenueByPaymentMethod(ctx)
	if err != nil {
		return nil, err
	}
	var revenueByPaymentMethodData []RevenueByPaymentMethod
	for _, paymentMethod := range revenueByPaymentMethod {
		convertedRevenue, err := convertToFloat64(paymentMethod.Revenue)
		if err != nil {
			return nil, err
		}
		revenueByPaymentMethodData = append(revenueByPaymentMethodData, RevenueByPaymentMethod{
			PaymentMethod: paymentMethod.PaymentMethod.String,
			Revenue:       convertedRevenue,
		})
	}

	// Get subscriptions over time data.
	subscriptionsOverTime, err := m.DB.SubscriptionsOverTime(ctx)
	if err != nil {
		return nil, err
	}
	var subscriptionsOverTimeData []SubscriptionOverTime
	for _, subscription := range subscriptionsOverTime {
		convertedSubscriptionsCount, err := convertToInt64(subscription.SubscriptionsCount)
		if err != nil {
			return nil, err
		}
		subscriptionsOverTimeData = append(subscriptionsOverTimeData, SubscriptionOverTime{
			Day:                subscription.Day,
			SubscriptionsCount: convertedSubscriptionsCount,
		})
	}

	// Get subscriptions by currency data.
	subscriptionsByCurrency, err := m.DB.SubscriptionsByCurrency(ctx)
	if err != nil {
		return nil, err
	}
	var subscriptionsByCurrencyData []SubscriptionsByCurrency
	for _, currency := range subscriptionsByCurrency {
		convertedCurrencyCount, err := convertToInt64(currency.CurrencyCount)
		if err != nil {
			return nil, err
		}
		subscriptionsByCurrencyData = append(subscriptionsByCurrencyData, SubscriptionsByCurrency{
			Currency:      currency.Currency.String,
			CurrencyCount: convertedCurrencyCount,
		})
	}

	// Build and return the multi-report.
	return &MultiReport{
		RevenueByPlan:           revenueByPlanData,
		ChallengedTransactions:  challengedTransactionsData,
		RevenueByPaymentMethod:  revenueByPaymentMethodData,
		SubscriptionsOverTime:   subscriptionsOverTimeData,
		SubscriptionsByCurrency: subscriptionsByCurrencyData,
	}, nil
}

// AdminGetSubscriptionStatsReports retrieves the overall subscription statistics report.
// It includes both single subscription statistics and multiple grouped reports.
func (m AdminModel) AdminGetSubscriptionStatsReports() (*SubscriptionStats, error) {
	// Create a context with a timeout for database operations.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Retrieve single subscription statistics.
	singleSubscriptionReport, err := m.getSingleReports(ctx)
	if err != nil {
		return nil, err
	}

	// Retrieve multi-report statistics.
	multiReport, err := m.getMultiReport(ctx)
	if err != nil {
		return nil, err
	}

	// Build and return the final subscription statistics report.
	subscriptionStats := SubscriptionStats{
		SingleSubscriptionReport: *singleSubscriptionReport,
		MultiReport:              *multiReport,
	}

	return &subscriptionStats, nil
}

// Helper function to safely convert interface{} to float64.
// Handles different types like float64 and []uint8 (from DB strings).
func convertToFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case []uint8: // if it's a string in the DB
		return strconv.ParseFloat(string(v), 64)
	default:
		return 0, fmt.Errorf("unexpected type for float64: %T", value)
	}
}

// Helper function to safely convert interface{} to int64.
// Handles different types like int64 and []uint8 (from DB strings).
func convertToInt64(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int64:
		return v, nil
	case []uint8: // if it's a string in the DB
		return strconv.ParseInt(string(v), 10, 64)
	default:
		return 0, fmt.Errorf("unexpected type for int64: %T", value)
	}
}