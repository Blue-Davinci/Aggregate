package data

import (
	"context"
	"database/sql"
	"errors"
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
)

var (
	ErrTransactionDeclined  = errors.New("transaction declined")
	ErrDuplicateTransaction = errors.New("duplicate transaction")
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
	User_ID     int64  `json:"-"`
	Plan_ID     int32  `json:"plan_id"`
	Amount      int64  `json:"amount"`
	Email       string `json:"email"`
	CallBackURL string `json:"callback_url"`
	Reference   string `json:"reference"`
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
	TransactionID      int64     `json:"transaction_id"`
	Payment_Method     string    `json:"payment_method"`
	Authorization_Code string    `json:"authorization_code"`
	Card_Last4         string    `json:"card_last4"`
	Card_Exp_Month     string    `json:"card_exp_month"`
	Card_Exp_Year      string    `json:"card_exp_year"`
	Card_Type          string    `json:"card_type"`
	Currency           string    `json:"currency"`
	Created_At         time.Time `json:"created_at"`
	Updated_At         time.Time `json:"updated_at"`
}

func ValidateTransactionData(v *validator.Validator, transactionData *TransactionData) {
	//amount
	v.Check(transactionData.Amount != 0, "amount", "must be varied")
	// plan id
	v.Check(transactionData.Plan_ID != 0, "plan_id", "must be provided")
}
func ValidateVerificationData(v *validator.Validator, transactionData *TransactionData) {
	//amount
	v.Check(transactionData.Reference != "", "reference", "must be varied")
	// plan id
	v.Check(transactionData.Plan_ID != 0, "plan_id", "must be provided")
}

func (m *PaymentsModel) InitializeTransaction() {

}
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
		Status:            payment_detail.Status,
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

func (m *PaymentsModel) GetPaymentPlanByID(plan_ID int32) (*Payment_Plan, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	plan, err := m.DB.GetPaymentPlanByID(ctx, plan_ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
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
