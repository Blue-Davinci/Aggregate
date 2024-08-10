package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

type PaymentOperation int8

const (
	PaymentOperationInitialize PaymentOperation = iota
	PaymentOperationVerify
	PaymentOperationRecurring
)

var (
	ErrTransactionDeclined           = errors.New("transaction declined")
	ErrDuplicateTransaction          = errors.New("duplicate transaction")
	ErrPaymentPlanNotFound           = errors.New("payment plan not found")
	ErrSubscriptionNotFound          = errors.New("subscription not found")
	ErrChallangedTransactionNotFound = errors.New("challanged transaction not found")
)

var (
	PaymentStatusActive    = "active"
	PaymentStatusRenewed   = "renewed"
	PaymentStatusExpired   = "expired"
	PaymentStatusCancelled = "cancelled"
	PaymentStatusSuccess   = "successful"
	PaymentStatusFailed    = "failed"
	PaymentStatusPending   = "pending"
)

type PaymentsModel struct {
	DB *database.Queries
}

type Log struct {
	StartTime int64     `json:"start_time"`
	TimeSpent int64     `json:"time_spent"`
	Attempts  int64     `json:"attempts"`
	Errors    int64     `json:"errors"`
	Success   bool      `json:"success"`
	Mobile    bool      `json:"mobile"`
	Input     []string  `json:"input"`
	History   []History `json:"history"`
}

type History struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Time    int64  `json:"time"`
}

type Authorization struct {
	AuthorizationCode string  `json:"authorization_code"`
	Bin               string  `json:"bin"`
	Last4             string  `json:"last4"`
	ExpMonth          string  `json:"exp_month"`
	ExpYear           string  `json:"exp_year"`
	Channel           string  `json:"channel"`
	CardType          string  `json:"card_type"`
	Bank              string  `json:"bank"`
	CountryCode       string  `json:"country_code"`
	Brand             string  `json:"brand"`
	Reusable          bool    `json:"reusable"`
	Signature         string  `json:"signature"`
	AccountName       *string `json:"account_name"`
}

type Customer struct {
	ID                       int64   `json:"id"`
	FirstName                *string `json:"first_name"`
	LastName                 *string `json:"last_name"`
	Email                    string  `json:"email"`
	CustomerCode             string  `json:"customer_code"`
	Phone                    *string `json:"phone"`
	Metadata                 *string `json:"metadata"`
	RiskAction               string  `json:"risk_action"`
	InternationalFormatPhone *string `json:"international_format_phone"`
}

type Data struct {
	ID                 int64                  `json:"id"`
	Authorization_url  string                 `json:"authorization_url"` //for challanged cards
	Paused             bool                   `json:"paused"`            //for challanged cards
	Domain             string                 `json:"domain"`
	Status             string                 `json:"status"`
	Reference          string                 `json:"reference"`
	ReceiptNumber      *string                `json:"receipt_number"`
	Amount             int64                  `json:"amount"`
	Message            *string                `json:"message"`
	GatewayResponse    string                 `json:"gateway_response"`
	PaidAt             string                 `json:"paid_at"`
	CreatedAt          string                 `json:"created_at"`
	Channel            string                 `json:"channel"`
	Currency           string                 `json:"currency"`
	IPAddress          string                 `json:"ip_address"`
	Metadata           string                 `json:"metadata"`
	Log                Log                    `json:"log"`
	Fees               int64                  `json:"fees"`
	FeesSplit          *string                `json:"fees_split"`
	Authorization      Authorization          `json:"authorization"`
	Customer           Customer               `json:"customer"`
	Plan               *string                `json:"plan"`
	Split              map[string]interface{} `json:"split"`
	OrderID            *string                `json:"order_id"`
	PaidAtFormatted    string                 `json:"paidAt"`
	CreatedAtFormatted string                 `json:"createdAt"`
	RequestedAmount    int64                  `json:"requested_amount"`
	POSTransactionData *string                `json:"pos_transaction_data"`
	Source             *string                `json:"source"`
	FeesBreakdown      *string                `json:"fees_breakdown"`
	Connect            *string                `json:"connect"`
	TransactionDate    string                 `json:"transaction_date"`
	PlanObject         map[string]interface{} `json:"plan_object"`
	Subaccount         map[string]interface{} `json:"subaccount"`
}

type UnifiedPaymentResponse struct {
	VerifyResponse     VerifyResponse     `json:"verify_response"`
	InitializeResponse InitializeResponse `json:"initialize_response"`
}

type VerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type InitializeRequest struct {
	Email       string `json:"email"`
	Amount      int    `json:"amount"`
	CallbackURL string `json:"callback_url"`
}

type InitializeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

type TransactionData struct {
	User_ID            int64  `json:"-"`
	Plan_ID            int32  `json:"plan_id"`
	Amount             int64  `json:"amount"`
	Email              string `json:"email"`
	CallBackURL        string `json:"callback_url"`
	Reference          string `json:"reference"`
	Authorization_Code string `json:"authorization_code"`
}

// Payment_Plan struct represents all the info we will
// return in relation to our subscription plans.
type Payment_Plan struct {
	ID          int32     `json:"id"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	Description string    `json:"description"`
	Duration    string    `json:"duration"`
	Price       int64     `json:"amount"`
	Features    []string  `json:"features"`
	Created_At  time.Time `json:"created_at"`
	Updated_At  time.Time `json:"updated_at"`
	Status      string    `json:"status"`
}

// Payment_Confirmation
// id, user_id, plan_id, start_date, end_date, status, transaction_id;
type Payment_Details struct {
	ID                 uuid.UUID `json:"id"`
	User_ID            int64     `json:"user_id"`
	Plan_ID            int32     `json:"plan_id"`
	Start_Date         time.Time `json:"start_date"`
	End_Date           time.Time `json:"end_date"`
	Price              int64     `json:"price"`
	Status             string    `json:"status"`
	TransactionID      int64     `json:"-"`
	Payment_Method     string    `json:"payment_method"`
	Authorization_Code string    `json:"-"`
	Card_Last4         string    `json:"card_last4"`
	Card_Exp_Month     string    `json:"-"`
	Card_Exp_Year      string    `json:"-"`
	Card_Type          string    `json:"card_type"`
	Currency           string    `json:"currency"`
	Created_At         time.Time `json:"created_at"`
	Updated_At         time.Time `json:"updated_at"`
}

// Payment History struct represents a user's payment history.
// And includes data about their subscription
type PaymentHistory struct {
	Subscription   Subscription `json:"subscription"`
	Plan_Name      string       `json:"plan_name"`
	Plan_Image     string       `json:"plan_image"`
	Plan_Duration  string       `json:"plan_duration"`
	Transaction_ID int64        `json:"transaction_id"`
	Payment_Method string       `json:"payment_method"`
	Card_Last4     string       `json:"card_last4"`
	Card_Exp_Month string       `json:"card_exp_month"`
	Card_Exp_Year  string       `json:"card_exp_year"`
	Card_Type      string       `json:"card_type"`
	Currency       string       `json:"currency"`
	Created_At     time.Time    `json:"created_at"`
}

// Subscription struct represents a user's subscription to a plan.
type Subscription struct {
	ID         uuid.UUID `json:"id"`
	User_ID    int64     `json:"user_id"`
	Plan_ID    int32     `json:"plan_id"`
	Start_Date time.Time `json:"start_date"`
	End_Date   time.Time `json:"end_date"`
	Price      int64     `json:"price"`
	Status     string    `json:"status"`
	Updated_At time.Time `json:"updated_at"`
}

type RecurringSubscription struct {
	Subscription       Subscription `json:"subscription"`
	Currency           string       `json:"currency"`
	User_ID            int64        `json:"user_id"`
	User_Name          string       `json:"user_name"`
	User_Email         string       `json:"user_email"`
	Authorization_Code string       `json:"authorization_code"`
}

type ChallengedTransaction struct {
	ID                       int64     `json:"id"`
	User_ID                  int64     `json:"user_id"`
	ReferencedSubscriptionID uuid.UUID `json:"referenced_subscription_id"`
	AuthorizationUrl         string    `json:"authorization_url"`
	Reference                string    `json:"reference"`
	Status                   string    `json:"status"`
	Created_At               time.Time `json:"created_at"`
	Updated_At               time.Time `json:"updated_at"`
}

// ValidateTransactionData will validate the initialization transaction data provided by the client.
func ValidateTransactionData(v *validator.Validator, transactionData *TransactionData) {
	//amount
	v.Check(transactionData.Amount != 0, "amount", "must be varied")
	// plan id
	v.Check(transactionData.Plan_ID != 0, "plan_id", "must be provided")
}

// ValidateVerificationData will validate the validation transaction data provided by the client.
func ValidateVerificationData(v *validator.Validator, transactionData *TransactionData) {
	//amount
	v.Check(transactionData.Reference != "", "reference", "must be varied")
	// plan id
	v.Check(transactionData.Plan_ID != 0, "plan_id", "must be provided")
}

func ValidateChallengedTransaction(v *validator.Validator, challengedTransaction *ChallengedTransaction) {
	// Check that the challengedID is provided
	v.Check(challengedTransaction.ID != 0, "ID", "must be valid or provided")
	// check status
	status := challengedTransaction.Status
	v.Check(status != "", "status", "must be provided")
	// check provided status is either data.pending or data.failed, data.successful
	v.Check(status == PaymentStatusPending || status == PaymentStatusFailed || status == PaymentStatusSuccess, "status", "must be a valid status")
}

func ValidateSubscriptionStatus(v *validator.Validator, subscription *Subscription) {
	_, isValid := ValidateUUID(subscription.ID.String())
	v.Check(isValid, "id", "must be a valid UUID")
	// check status
	status := subscription.Status
	v.Check(status != "", "status", "must be provided")
	// check provided status is either data.pending or data.failed, data.successful
	v.Check(status == PaymentStatusCancelled, "status", "must be a valid status")
}

// GetSubscriptionByID will return a user's current subscription if it exists.
// We take in a user's ID and return a pointer to a Subscription and an error.
func (m *PaymentsModel) GetSubscriptionByID(userID int64) (*Subscription, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	subscription, err := m.DB.GetSubscriptionByID(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrSubscriptionNotFound
		default:
			return nil, err
		}
	}
	var userSub Subscription
	userSub.ID = subscription.ID
	userSub.User_ID = subscription.UserID
	userSub.Plan_ID = subscription.PlanID
	userSub.Start_Date = subscription.StartDate
	userSub.End_Date = subscription.EndDate
	priceStr := subscription.Price
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, err
	}
	userSub.Status = subscription.Status
	userSub.Price = int64(price)
	// we're good, we return the subscription
	return &userSub, nil
}

// CreateSubscription will create a new subscription for a user.
// This takes in payment details provided from the client and is only activated when
// the payment is successful. We return an error if the transaction fails.
// We also return an error if we detect a pre-existing transaction/auth code
// which denotes an already existing and active subscription.
func (m *PaymentsModel) CreateSubscription(payment_detail *Payment_Details) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	priceStr := strconv.FormatFloat(float64(payment_detail.Price), 'f', 2, 64)
	queyresult, err := m.DB.CreateSubscription(ctx, database.CreateSubscriptionParams{
		UserID:            payment_detail.User_ID,
		PlanID:            payment_detail.Plan_ID,
		StartDate:         payment_detail.Start_Date,
		EndDate:           payment_detail.End_Date,
		Price:             priceStr,
		Status:            PaymentStatusActive, // set it to active
		TransactionID:     payment_detail.TransactionID,
		PaymentMethod:     sql.NullString{String: payment_detail.Payment_Method, Valid: payment_detail.Payment_Method != ""},
		AuthorizationCode: sql.NullString{String: payment_detail.Authorization_Code, Valid: payment_detail.Authorization_Code != ""},
		CardLast4:         sql.NullString{String: payment_detail.Card_Last4, Valid: payment_detail.Card_Last4 != ""},
		CardExpMonth:      sql.NullString{String: payment_detail.Card_Exp_Month, Valid: payment_detail.Card_Exp_Month != ""},
		CardExpYear:       sql.NullString{String: payment_detail.Card_Exp_Year, Valid: payment_detail.Card_Exp_Year != ""},
		CardType:          sql.NullString{String: payment_detail.Card_Type, Valid: payment_detail.Card_Type != ""},
		Currency:          sql.NullString{String: payment_detail.Currency, Valid: payment_detail.Currency != ""},
	})
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "subscriptions_transaction_id_key"` ||
			err.Error() == `pq: duplicate key value violates unique constraint "subscriptions_transaction_id_authorization_code_key"`:
			return ErrDuplicateTransaction
		default:
			return err
		}
	}
	// now fill in the payment_detail missing fields with our returned data
	payment_detail.ID = queyresult.ID
	payment_detail.Created_At = queyresult.CreatedAt
	payment_detail.Updated_At = queyresult.UpdatedAt
	// we are okay so we return nil
	return nil
}

func (m *PaymentsModel) CreateChallangedTransaction(subscription *RecurringSubscription, url, challanged_error, reference string) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.DB.CreateChallangedTransaction(ctx, database.CreateChallangedTransactionParams{
		UserID:                   subscription.User_ID,
		ReferencedSubscriptionID: subscription.Subscription.ID,
		AuthorizationUrl:         url,
		Reference:                reference,
		Status:                   PaymentStatusPending,
	})
	if err != nil {
		return err
	}
	return nil
}

// CreateFailedTransaction will create a failed transaction in the database. This failed transactions will be used
// by the admin to check the failed transactions and take necessary actions.
func (m *PaymentsModel) CreateFailedTransaction(paymentDetails *Payment_Details, failure_reason, reference, error_code string) (int64, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	priceStr := strconv.FormatFloat(float64(paymentDetails.Price), 'f', 2, 64)
	queryResult, err := m.DB.CreateFailedTransaction(ctx, database.CreateFailedTransactionParams{
		UserID:            paymentDetails.User_ID,
		SubscriptionID:    paymentDetails.ID,
		AuthorizationCode: sql.NullString{String: paymentDetails.Authorization_Code, Valid: true},
		Reference:         reference,
		Amount:            priceStr,
		FailureReason:     sql.NullString{String: failure_reason, Valid: true},
		ErrorCode:         sql.NullString{String: error_code, Valid: true},
		CardLast4:         sql.NullString{String: paymentDetails.Card_Last4, Valid: true},
		CardExpMonth:      sql.NullString{String: paymentDetails.Card_Exp_Month, Valid: true},
		CardExpYear:       sql.NullString{String: paymentDetails.Card_Exp_Year, Valid: true},
		CardType:          sql.NullString{String: paymentDetails.Card_Type, Valid: true},
	})
	if err != nil {
		return 0, err
	}
	return queryResult.ID, nil
}

func (m *PaymentsModel) GetPendingChallengedTransactionsByUser(userID int64) ([]*ChallengedTransaction, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.GetPendingChallengedTransactionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Initialize the slice as an empty slice
	challengedTransactions := []*ChallengedTransaction{}

	for _, row := range rows {
		challengedTransaction := &ChallengedTransaction{
			ID:                       row.ID,
			User_ID:                  row.UserID,
			ReferencedSubscriptionID: row.ReferencedSubscriptionID,
			AuthorizationUrl:         row.AuthorizationUrl,
			Reference:                row.Reference,
			Status:                   row.Status,
			Created_At:               row.CreatedAt,
			Updated_At:               row.UpdatedAt,
		}
		challengedTransactions = append(challengedTransactions, challengedTransaction)
	}
	return challengedTransactions, nil
}

// GetAllSubscriptionsByID() returns the subscription history of a user.
// It also supports pagination and metadata reporting.
func (m *PaymentsModel) GetAllSubscriptionsByID(userID int64, filters Filters) ([]*PaymentHistory, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := m.DB.GetAllSubscriptionsByID(ctx, database.GetAllSubscriptionsByIDParams{
		UserID: userID,
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	if err != nil {
		return nil, Metadata{}, err
	}
	totalRecords := 0
	payment_histories := []*PaymentHistory{}
	for _, row := range rows {
		var payment_history PaymentHistory
		totalRecords = int(row.TotalRecords)
		payment_history.Transaction_ID = row.TransactionID
		payment_history.Payment_Method = row.PaymentMethod.String
		payment_history.Card_Last4 = row.CardLast4.String
		payment_history.Card_Exp_Month = row.CardExpMonth.String
		payment_history.Card_Exp_Year = row.CardExpYear.String
		payment_history.Card_Type = row.CardType.String
		payment_history.Currency = row.Currency.String
		payment_history.Created_At = row.CreatedAt
		// plan details
		payment_history.Plan_Name = row.PlanName
		payment_history.Plan_Image = row.PlanImage
		payment_history.Plan_Duration = row.PlanDuration
		// fill in the subscription details
		var subscription Subscription
		subscription.ID = row.ID
		subscription.User_ID = row.UserID
		subscription.Plan_ID = row.PlanID
		subscription.Start_Date = row.StartDate
		subscription.End_Date = row.EndDate
		priceStr := row.Price
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, Metadata{}, err
		}
		subscription.Price = int64(price)
		subscription.Status = row.Status
		payment_history.Subscription = subscription
		payment_histories = append(payment_histories, &payment_history)
	}
	fmt.Println("Total rows: ", totalRecords)
	// calculate the metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return payment_histories, metadata, nil
}

// GetAllExpiredSubscriptions() returns all the expired subscriptions.
// The function supports pagination and metadata reporting.
// It works in tandem with our expiration checker checking for all items
// whose dates have passed and are expired
func (m *PaymentsModel) GetAllExpiredSubscriptions(filters Filters) ([]*RecurringSubscription, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := m.DB.GetAllExpiredSubscriptions(ctx, database.GetAllExpiredSubscriptionsParams{
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	if err != nil {
		fmt.Println("Error: ", err)
		return nil, Metadata{}, err
	}
	recurring_subscriptions := []*RecurringSubscription{}
	totalRecords := 0
	for _, row := range rows {
		var recurring_subscription RecurringSubscription
		totalRecords = int(row.TotalRecords)
		recurring_subscription.User_ID = row.UserID
		recurring_subscription.User_Name = row.Name
		recurring_subscription.User_Email = row.Email
		recurring_subscription.Authorization_Code = row.AuthorizationCode.String
		recurring_subscription.Currency = row.Currency.String
		// fill in the subscription details
		var subscription Subscription
		subscription.ID = row.SubscriptionID
		subscription.Plan_ID = row.PlanID
		priceStr, err := strconv.ParseFloat(row.Price, 64)
		if err != nil {
			return nil, Metadata{}, err
		}
		subscription.Price = int64(priceStr)
		// Add the subscription to the recurring_subscription
		recurring_subscription.Subscription = subscription
		// fill in the subscription details
		recurring_subscriptions = append(recurring_subscriptions, &recurring_subscription)
	}
	// calculate the metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return recurring_subscriptions, metadata, nil
}

// GetPaymentDetailsByID will return an individual plan giving back all
// available information about the plan.
func (m *PaymentsModel) GetPaymentPlanByID(plan_ID int32) (*Payment_Plan, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	plan, err := m.DB.GetPaymentPlanByID(ctx, plan_ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrPaymentPlanNotFound
		default:
			return nil, err
		}
	}
	var payment_plan Payment_Plan
	payment_plan.ID = plan.ID
	payment_plan.Name = plan.Name
	payment_plan.Image = plan.Image
	payment_plan.Description = plan.Description.String
	payment_plan.Duration = plan.Duration
	priceStr := plan.Price
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return nil, err
	}
	payment_plan.Price = int64(price)
	payment_plan.Features = plan.Features
	payment_plan.Created_At = plan.CreatedAt
	payment_plan.Updated_At = plan.UpdatedAt
	payment_plan.Status = plan.Status
	// we're good, we return the payment_plan
	return &payment_plan, nil
}

// GetPaymentPlans will return all the available payment plans.
// It's a simple check and return.
func (m *PaymentsModel) GetPaymentPlans() ([]*Payment_Plan, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.GetPaymentPlans(ctx)
	if err != nil {
		return nil, err
	}
	payment_plans := []*Payment_Plan{}
	for _, row := range rows {
		var payment_plan Payment_Plan
		payment_plan.ID = row.ID
		payment_plan.Name = row.Name
		payment_plan.Image = row.Image
		payment_plan.Description = row.Description.String
		payment_plan.Duration = row.Duration
		priceStr := row.Price
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, err
		}
		payment_plan.Price = int64(price)
		payment_plan.Features = row.Features
		payment_plan.Created_At = row.CreatedAt
		payment_plan.Updated_At = row.UpdatedAt
		payment_plan.Status = row.Status

		payment_plans = append(payment_plans, &payment_plan)
	}
	return payment_plans, nil
}

// GetPendingChallengedTransactionBySubscriptionID() gets a pending challenged transaction
// specific to a particulat user subscription.
// We do this to prevent multiple transactions from being created for the same subscription.
// especially when or during a recurring subscription charge
func (m *PaymentsModel) GetPendingChallengedTransactionBySubscriptionID(SubscriptionID uuid.UUID) (*ChallengedTransaction, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	queryResult, err := m.DB.GetPendingChallengedTransactionBySubscriptionID(ctx, SubscriptionID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrChallangedTransactionNotFound
		default:
			return nil, err
		}
	}
	challengedSubscription := &ChallengedTransaction{
		ID:                       queryResult.ID,
		User_ID:                  queryResult.UserID,
		ReferencedSubscriptionID: queryResult.ReferencedSubscriptionID,
		AuthorizationUrl:         queryResult.AuthorizationUrl,
		Reference:                queryResult.Reference,
		Status:                   queryResult.Status,
		Created_At:               queryResult.CreatedAt,
		Updated_At:               queryResult.UpdatedAt,
	}
	return challengedSubscription, nil
}

// UpdateSubscriptionStatusAfterRenewal will update the status of an immediate subscription
// That is, if we were dealing with subscription x, and we were re-charging it and the
// recharge/recurring charge was successful, this function will update the status from
// being "expired" to "renewed".
// This will help us keep track of finished transactions in the table. Also it will prevent
// the subscription from being recharged again when the checker job runs.
func (m *PaymentsModel) UpdateSubscriptionStatus(subscriptionID uuid.UUID, status string, userID int64) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// will ignore the result
	_, err := m.DB.UpdateSubscriptionStatus(ctx, database.UpdateSubscriptionStatusParams{
		ID:     subscriptionID,
		Status: status,
		UserID: userID,
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateSubscriptionStatusAfterExpiration will update the status of all subscriptions
// setting the ones that are expired to the status "expired".
func (m *PaymentsModel) UpdateSubscriptionStatusAfterExpiration() ([]*Subscription, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// will ignore the result
	rows, err := m.DB.UpdateSubscriptionStatusAfterExpiration(ctx)
	if err != nil {
		return nil, err
	}
	var subscriptions []*Subscription
	for _, row := range rows {
		subscription := &Subscription{
			ID:         row.ID,
			User_ID:    row.UserID,
			Updated_At: row.UpdatedAt,
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}

// UpdateChallengedTransactionStatus will update the status of a challenged transaction based on
// whether it was successful or not. This function should only be hit when there is a subscription
// that has been challenged receiving its ID and the status to set.
func (m *PaymentsModel) UpdateChallengedTransactionStatus(challengedSubscriptionID, user_id int64, status string) (*ChallengedTransaction, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// will ignore the result
	queryResult, err := m.DB.UpdateChallengedTransactionStatus(ctx, database.UpdateChallengedTransactionStatusParams{
		ID:     challengedSubscriptionID,
		Status: status,
		UserID: user_id,
	})
	if err != nil {
		return nil, err
	}
	// get the updated challenged transaction
	challengedSubscription := &ChallengedTransaction{
		ID:                       queryResult.ID,
		User_ID:                  queryResult.UserID,
		ReferencedSubscriptionID: queryResult.ReferencedSubscriptionID,
		AuthorizationUrl:         queryResult.AuthorizationUrl,
		Reference:                queryResult.Reference,
		Status:                   queryResult.Status,
		Created_At:               queryResult.CreatedAt,
		Updated_At:               queryResult.UpdatedAt,
	}

	return challengedSubscription, nil
}
