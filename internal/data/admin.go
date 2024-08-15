package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
)

type AdminModel struct {
	DB *database.Queries
}

var (
	ErrDuplicatePaymentPlan = errors.New("duplicate payment plan")
)

// Represents a user struct, with the permissions field being an array of strings.
type AdminUser struct {
	ID          int64       `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Password    string      `json:"-"` // we omit this yk, Even for godmode we're decent people (not that it'll matter anyway)
	Activated   bool        `json:"activated"`
	Version     int         `json:"version"`
	User_Img    string      `json:"user_img"`
	Permissions Permissions `json:"permissions"`
}

// Admin statistics represents the statistics for the admin page.
type AdminStatistics struct {
	UserStatistics         UserStatistics         `json:"user_statistics"`
	SubscriptionStatistics SubscriptionStatistics `json:"subscription_statistics"`
	CommentStatistics      CommentStatistics      `json:"comment_statistics"`
}

// UserStatistics represents the statistics for the users.
type UserStatistics struct {
	Total_Users   int64   `json:"total_users"`
	Active_Users  int64   `json:"active_users"`
	New_Signups   int64   `json:"new_signups"`
	Total_Revenue float64 `json:"total_revenue"`
}

// SubscriptionStatistics represents the statistics for the subscriptions.
type SubscriptionStatistics struct {
	Total_Revenue            float64 `json:"total_revenue"`
	Active_Subscriptions     int64   `json:"active_subscriptions"`
	Cancelled_Subscriptions  int64   `json:"cancelled_subscriptions"`
	Expired_Subscriptions    int64   `json:"expired_subscriptions"`
	Most_Used_Payment_Method string  `json:"most_used_payment_method"`
}

// CommentStatistics represents the statistics for the comments.
type CommentStatistics struct {
	Total_Comments  int64 `json:"total_comments"`
	Recent_Comments int64 `json:"recent_comments"`
}

// AdminGetAllUsers() returns all available users in the DB. This route supports a full text search for the user Name as well
func (m *AdminModel) AdminGetAllUsers(nameQuery string, filters Filters) ([]*AdminUser, Metadata, error) {
	// Create a context with a timeout of 3 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Call the GetAllUsers method from the database package.
	dbUsers, err := m.DB.AdminGetAllUsers(ctx, database.AdminGetAllUsersParams{
		Column1: nameQuery,
		Limit:   int32(filters.limit()),
		Offset:  int32(filters.offset()),
	})
	if err != nil {
		return nil, Metadata{}, err
	}

	// Make an array of pointers to AdminUser.
	var users []*AdminUser
	totalRecords := 0
	// Iterate over the returned users and append them to the users slice.
	// Iterate over the returned users and append them to the users slice.
	for _, dbUser := range dbUsers {
		totalRecords = int(dbUser.TotalCount)

		// Convert []uint8 to string
		rawPermissions := string(dbUser.Permissions.([]uint8))

		// Clean up the string (e.g., remove surrounding curly braces)
		rawPermissions = strings.Trim(rawPermissions, "{}")

		// Split the string into individual permissions
		userPermissions := strings.Split(rawPermissions, ",")

		// Handle the case where the result is an empty string
		if len(userPermissions) == 1 && userPermissions[0] == "" {
			userPermissions = []string{"Default User"}
		}

		adminUser := &AdminUser{
			ID:          dbUser.ID,
			CreatedAt:   dbUser.CreatedAt,
			Name:        dbUser.Name,
			Email:       dbUser.Email,
			Activated:   dbUser.Activated,
			Version:     int(dbUser.Version),
			User_Img:    dbUser.UserImg,
			Permissions: userPermissions,
		}
		users = append(users, adminUser)
	}
	// calculate metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Return the users slice, metadata and a nil error.
	return users, metadata, nil
}

func (m *PaymentsModel) AdminGetPaymentPlans() ([]*Payment_Plan, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.AdminGetAllPaymentPlans(ctx)
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
		payment_plan.Version = row.Version

		payment_plans = append(payment_plans, &payment_plan)
	}
	return payment_plans, nil
}

// AdminGetPaymentPlanByID
func (m *PaymentsModel) AdminGetPaymentPlanByID(planID int32) (*Payment_Plan, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	plan, err := m.DB.AdminGetPaymentPlanByID(ctx, planID)
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
	payment_plan.Version = plan.Version
	// we're good, we return the payment_plan
	return &payment_plan, nil
}

// AdminGetAllSubscriptions() returns all the subscriptions in the database.
func (m *PaymentsModel) AdminGetAllSubscriptions(filters Filters) ([]*PaymentHistory, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := m.DB.AdminGetAllSubscriptions(ctx, database.AdminGetAllSubscriptionsParams{
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

// AdminGetStatistics() returns all the statistics, aggregated together for representation in the frontend.
func (m *AdminModel) AdminGetStatistics() (*AdminStatistics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Call the GetStatistics method from the database package.
	queryResult, err := m.DB.AdminGetStatistics(ctx)
	if err != nil {
		return nil, err
	}
	// Type assertion for TotalRevenue
	totalRevenue, err := strconv.ParseFloat(queryResult.TotalRevenue, 64)
	if err != nil {
		fmt.Printf("unexpected type for TotalRevenue: %T\n", queryResult.TotalRevenue)
		return nil, errors.New("unexpected type for TotalRevenue")
	}
	// Type assertion for MostUsedPaymentMethod
	mostUsedPaymentMethod, ok := queryResult.MostUsedPaymentMethod.(string)
	if !ok {
		fmt.Printf("unexpected type for MostUsedPaymentMethod: %T\n", queryResult.MostUsedPaymentMethod)
		return nil, errors.New("unexpected type for MostUsedPaymentMethod")
	}
	//Type assertion for Total

	// Create a new AdminStatistics struct instance.
	adminStatistics := &AdminStatistics{
		UserStatistics: UserStatistics{
			Total_Users:  queryResult.TotalUsers,
			Active_Users: queryResult.ActiveUsers,
			New_Signups:  queryResult.NewSignups,
		},
		SubscriptionStatistics: SubscriptionStatistics{
			Total_Revenue:            totalRevenue,
			Active_Subscriptions:     queryResult.ActiveSubscriptions,
			Cancelled_Subscriptions:  queryResult.CancelledSubscriptions,
			Expired_Subscriptions:    queryResult.ExpiredSubscriptions,
			Most_Used_Payment_Method: mostUsedPaymentMethod,
		},
		CommentStatistics: CommentStatistics{
			Total_Comments:  queryResult.TotalComments,
			Recent_Comments: queryResult.RecentComments,
		},
	}

	return adminStatistics, nil
}

// AdminCreatePaymentPlans() creates a new payment plan in the database.
func (m *PaymentsModel) AdminCreatePaymentPlans(paymentPlan *Payment_Plan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queryResult, err := m.DB.AdminCreatePaymentPlan(ctx, database.AdminCreatePaymentPlanParams{
		Name:        paymentPlan.Name,
		Image:       paymentPlan.Image,
		Description: sql.NullString{String: paymentPlan.Description, Valid: paymentPlan.Description != ""},
		Duration:    paymentPlan.Duration,
		Price:       strconv.FormatFloat(float64(paymentPlan.Price), 'f', 2, 64),
		Features:    paymentPlan.Features,
		Status:      paymentPlan.Status,
	})
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "unique_name"`:
			return ErrDuplicatePaymentPlan
		default:
			return err
		}
	}
	paymentPlan.ID = queryResult.ID
	paymentPlan.Created_At = queryResult.CreatedAt
	paymentPlan.Updated_At = queryResult.UpdatedAt

	return nil
}

func (m *PaymentsModel) AdminUpdatePaymentPlan(paymentPlan *Payment_Plan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	/*
			    Name        string
		    Image       string
		    Description sql.NullString
		    Duration    string
		    Price       string
		    Features    []string
		    Status      string
		    ID          int32
		    Version     int32
	*/
	// update the payment plan
	// we also include the version to prevent edit conflicts
	version, err := m.DB.AdminUpdatePaymentPlan(ctx, database.AdminUpdatePaymentPlanParams{
		ID:          paymentPlan.ID,
		Name:        paymentPlan.Name,
		Image:       paymentPlan.Image,
		Description: sql.NullString{String: paymentPlan.Description, Valid: paymentPlan.Description != ""},
		Duration:    paymentPlan.Duration,
		Price:       strconv.FormatFloat(float64(paymentPlan.Price), 'f', 2, 64),
		Features:    paymentPlan.Features,
		Status:      paymentPlan.Status,
		Version:     paymentPlan.Version,
	})
	// check for an edit conflict, if there was, we return it specifically.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	// update the version
	paymentPlan.Version = version
	// payment plan was updated successfully
	return nil
}
