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
	"github.com/google/uuid"
)

type AdminModel struct {
	DB *database.Queries
}

var (
	ErrDuplicatePaymentPlan = errors.New("duplicate payment plan")
)

type AdminPaymentHistory struct {
	PaymentHistory              PaymentHistory `json:"payment_history"`
	Has_Challenged_Transactions bool           `json:"has_challenged_transactions"`
}

type AdminChallangedTransaction struct {
	ChallengedTransaction ChallengedTransaction `json:"challenged_transaction"`
	PlanID                int32                 `json:"plan_id"`
	PlanPrice             int64                 `json:"plan_price"`
	Start_Date            time.Time             `json:"start_date"`
	End_Date              time.Time             `json:"end_date"`
	UserName              string                `json:"user_name"`
	UserEmail             string                `json:"user_email"`
	UserImg               string                `json:"user_img"`
}

type SuperUsers struct {
	UserID          int64           `json:"user_id"`
	Name            string          `json:"name"`
	User_Img        string          `json:"user_img"`
	AdminPermission AdminPermission `json:"admin_permission"`
}

type AdminPermission struct {
	Permission_ID   int64  `json:"permission_id"`
	Permission_Code string `json:"permission_code"`
}

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

type AdminFeed struct {
	Feed            Feed          `json:"feed"`
	AdminFeedUser   AdminFeedUser `json:"feed_user"`
	Approval_Status string        `json:"approval_status"`
	Priority        string        `json:"priority"`
	FollowCount     int64         `json:"follow_count"`
}

type AdminFeedUser struct {
	UserID   int64  `json:"user_id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	User_Img string `json:"user_img"`
}

type AdminFeedStats struct {
	AdminPendingFeedStats AdminPendingFeedStats `json:"admin_pending_feed_stats"`
	TotalHiddenFeeds      int64                 `json:"total_hidden_feeds"`
	MostCommonFeedType    string                `json:"most_common_feed_type"`
}

type AdminPendingFeedStats struct {
	TotalPendingFeeds  int64 `json:"total_pending_feeds"`
	TotalApprovedFeeds int64 `json:"total_approved_feeds"`
	TotalRejectedFeeds int64 `json:"total_rejected_feeds"`
	OldestPendingFeed  int64 `json:"oldest_pending_feed"`
}

func UpdateAdminFeedFields(input *AdminFeedInput, adminFeed *AdminFeed) {
	// if name, url, imgurl,feedtype etc
	if input.Name != nil {
		adminFeed.Feed.Name = *input.Name
	}
	if input.Url != nil {
		adminFeed.Feed.Url = *input.Url
	}
	if input.ImgURL != nil {
		adminFeed.Feed.ImgURL = *input.ImgURL
	}
	if input.FeedType != nil {
		adminFeed.Feed.FeedType = *input.FeedType
	}
	if input.FeedDescription != nil {
		adminFeed.Feed.FeedDescription = *input.FeedDescription
	}
	if input.Is_Hidden != nil {
		adminFeed.Feed.Is_Hidden = *input.Is_Hidden
	}
	if input.ApprovalStatus != nil {
		adminFeed.Approval_Status = *input.ApprovalStatus
	}
	if input.Priority != nil {
		adminFeed.Priority = *input.Priority
	}
}

// AdminGetAllUsers() returns all available users in the DB. This route supports a full text search for the user Name as well
func (m AdminModel) AdminGetAllUsers(nameQuery string, filters Filters) ([]*AdminUser, Metadata, error) {
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

func (m AdminModel) AdminGetPaymentPlans() ([]*Payment_Plan, error) {
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
func (m AdminModel) AdminGetPaymentPlanByID(planID int32) (*Payment_Plan, error) {
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

// GetAllSuperUsersWithPermissions() is an admin method that retrieves all super users
// currently this method will return admins/moderators since they are the only individuals
// with permissions in the system. But if and when the system is expanded to include other
// permissions such as  "comment:write" for example to ban users from commenting, this
// method will need to be updated to include/leave out those user.
func (m AdminModel) AdminGetAllSuperUsersWithPermissions() ([]*SuperUsers, error) {
	// set up context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// create our permissions
	var superUsers []*SuperUsers
	// call the database method
	dbSuperUsers, err := m.DB.GetAllSuperUsersWithPermissions(ctx)
	if err != nil {
		return nil, err
	}
	for _, superUser := range dbSuperUsers {
		superUsers = append(superUsers, &SuperUsers{
			UserID:   superUser.UserID,
			Name:     superUser.Name,
			User_Img: superUser.UserImg,
			AdminPermission: AdminPermission{
				Permission_ID:   superUser.PermissionID,
				Permission_Code: superUser.PermissionCode,
			},
		})
	}
	// return permissions
	return superUsers, nil
}

// AdminGetAllSubscriptions() returns all the subscriptions in the database.
func (m AdminModel) AdminGetAllSubscriptions(filters Filters) ([]*AdminPaymentHistory, Metadata, error) {
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
	admin_payment_histories := []*AdminPaymentHistory{}
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
		// type asser
		// aggregate the payment history
		admin_payment := &AdminPaymentHistory{
			PaymentHistory:              payment_history,
			Has_Challenged_Transactions: row.HasChallengedTransactions,
		}
		admin_payment_histories = append(admin_payment_histories, admin_payment)
	}
	//fmt.Println("Total rows: ", totalRecords)
	// calculate the metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return admin_payment_histories, metadata, nil
}

// AdminGetChallaengedTransactionsBySubscriptionID() returns all the challenged transactions for a specific subscription.
// the item includes the infor for the challenged transaction alongside plan details, user details.
// The search is by the subscription ID
func (m AdminModel) AdminGetChallaengedTransactionsBySubscriptionID(subscriptionID uuid.UUID) ([]*AdminChallangedTransaction, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rows, err := m.DB.AdminGetChallengedTransactionsBySubscriptionID(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	admin_challenged_transactions := []*AdminChallangedTransaction{}
	for _, row := range rows {
		var admin_challenged_transaction AdminChallangedTransaction
		var challenged_transaction ChallengedTransaction
		challenged_transaction.ID = row.TransactionID
		challenged_transaction.User_ID = row.UserID
		challenged_transaction.ReferencedSubscriptionID = row.ReferencedSubscriptionID
		challenged_transaction.Reference = row.Reference
		challenged_transaction.AuthorizationUrl = row.AuthorizationUrl
		challenged_transaction.Created_At = row.CreatedAt
		challenged_transaction.Updated_At = row.UpdatedAt
		challenged_transaction.Status = row.Status

		// add additional info
		admin_challenged_transaction.ChallengedTransaction = challenged_transaction
		admin_challenged_transaction.PlanID = row.PlanID
		priceStr := row.Price
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			return nil, err
		}
		admin_challenged_transaction.PlanPrice = int64(price)
		admin_challenged_transaction.Start_Date = row.StartDate
		admin_challenged_transaction.End_Date = row.EndDate
		// user data
		admin_challenged_transaction.UserName = row.UserName
		admin_challenged_transaction.UserEmail = row.UserEmail
		admin_challenged_transaction.UserImg = row.UserImg
		// append to the list
		admin_challenged_transactions = append(admin_challenged_transactions, &admin_challenged_transaction)
	}
	return admin_challenged_transactions, nil
}

// AdminGetStatistics() returns all the statistics, aggregated together for representation in the frontend.
// Currently gets subscription, content and user statistics
func (m AdminModel) AdminGetStatistics() (*AdminStatistics, error) {
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
func (m AdminModel) AdminCreatePaymentPlans(paymentPlan *Payment_Plan) error {
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

// AdminUpdatePaymentPlan() updates a payment plan in the database.
func (m AdminModel) AdminUpdatePaymentPlan(paymentPlan *Payment_Plan) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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

// AdminUpdatePermissionCode() updates the permission code for a specific permission.
func (m AdminModel) AdminUpdatePermissionCode(permission AdminPermission) (*AdminPermission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// update the permission code
	_, err := m.DB.AdminUpdatePermissionCode(ctx, database.AdminUpdatePermissionCodeParams{
		ID:   permission.Permission_ID,
		Code: permission.Permission_Code,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrPermissionNotFound
		default:
			return nil, err
		}
	}
	// no issues, we return the permission
	return &permission, nil
}

// AdminCreateNewPermission() creates a new permission in the database.
// This is a permission type that can be assigned to a user.
// For example, a user can have the permission "admin:write" which grants
// an admin the write permission to the admin section of the application.
func (m AdminModel) AdminCreateNewPermission(permission string) (*AdminPermission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dbPermission, err := m.DB.AdminCreateNewPermission(ctx, permission)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "unique_code"`:
			return nil, ErrDuplicatePermission
		default:
			return nil, err
		}
	}
	adminPermission := &AdminPermission{
		Permission_ID:   dbPermission.ID,
		Permission_Code: dbPermission.Code,
	}
	return adminPermission, nil
}

// AdminDeletePermission() deletes an exact permission from our list of
// available permissions
func (m AdminModel) AdminDeletePermission(permissionID int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// perform the deletion
	_, err := m.DB.AdminDeletePermission(ctx, permissionID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrPermissionNotFound
		default:
			return err
		}
	}
	// no issues, we return nil
	return nil
}

func (m AdminModel) AdminGetFeedsPendingApproval(name, feed_type string, filters Filters) ([]*AdminFeed, Metadata, AdminPendingFeedStats, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.AdminGetFeedsPendingApproval(ctx, database.AdminGetFeedsPendingApprovalParams{
		Column1:  name,
		FeedType: feed_type, // Convert string to sql.NullString
		Limit:    int32(filters.limit()),
		Offset:   int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, AdminPendingFeedStats{}, err
	}
	var adminPendingFeedStats AdminPendingFeedStats
	totalRecords := 0
	adminFeeds := []*AdminFeed{}
	for _, row := range rows {
		// create a new feed
		var feed Feed
		totalRecords = int(row.TotalCount)
		feed.ID = row.ID
		feed.CreatedAt = row.CreatedAt
		feed.UpdatedAt = row.UpdatedAt
		feed.Name = row.Name
		feed.Url = row.Url
		feed.Version = row.Version
		feed.UserID = row.UserID.Int64
		feed.ImgURL = row.ImgUrl
		feed.FeedType = row.FeedType
		feed.FeedDescription = row.FeedDescription
		feed.Is_Hidden = row.IsHidden
		// create a new admin feed
		adminFeed := &AdminFeed{
			Feed:            feed,
			AdminFeedUser:   AdminFeedUser{UserID: row.UserID.Int64, Name: row.UserName.String, Email: row.UserEmail.String, User_Img: row.UserImg},
			Approval_Status: row.ApprovalStatus,
			Priority:        row.Priority.String,
		}
		// fill statistics
		adminPendingFeedStats.TotalPendingFeeds = row.TotalPendingFeeds.Int64
		adminPendingFeedStats.TotalApprovedFeeds = row.TotalApprovedFeeds.Int64
		adminPendingFeedStats.TotalRejectedFeeds = row.TotalRejectedFeeds.Int64
		adminPendingFeedStats.OldestPendingFeed = row.OldestPendingDays.Int64
		// append to the list
		adminFeeds = append(adminFeeds, adminFeed)
	}
	// calculate the metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return adminFeeds, metadata, adminPendingFeedStats, nil
}

func (m AdminModel) AdminGetFeedByID(feedID uuid.UUID) (*AdminFeed, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve the feed
	row, err := m.DB.GetFeedById(ctx, feedID)
	// check for an error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// create a new feed
	var feed Feed
	feed.ID = row.ID
	feed.CreatedAt = row.CreatedAt
	feed.UpdatedAt = row.UpdatedAt
	feed.Name = row.Name
	feed.Url = row.Url
	feed.Version = row.Version
	feed.UserID = row.UserID
	feed.ImgURL = row.ImgUrl
	feed.FeedType = row.FeedType
	feed.FeedDescription = row.FeedDescription
	feed.Is_Hidden = row.IsHidden
	// create a new admin feed
	adminFeed := &AdminFeed{
		Feed:            feed,
		AdminFeedUser:   AdminFeedUser{UserID: row.UserID, Name: row.Name, User_Img: row.ImgUrl},
		Approval_Status: row.ApprovalStatus,
		Priority:        row.Priority,
	}
	// return the feed
	return adminFeed, nil
}

// AdminGetAllFeedsWithStatistics() returns all the feeds in the database with their statistics.
// The statistics include the total number of feeds, the total number of pending feeds, the total number of approved feeds,
// the total number of rejected feeds, the total number of hidden feeds, and the most common feed type.
// This endpoint supports both filters and pagination.
func (m AdminModel) AdminGetAllFeedsWithStatistics(name, feed_type string, filters Filters) ([]*AdminFeed, Metadata, AdminFeedStats, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.AdminGetAllFeedsWithStatistics(ctx, database.AdminGetAllFeedsWithStatisticsParams{
		Column1:  name,
		FeedType: feed_type, // Convert string to sql.NullString
		Limit:    int32(filters.limit()),
		Offset:   int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, AdminFeedStats{}, err
	}
	totalRecords := 0
	adminFeeds := []*AdminFeed{}
	var adminFeedStats AdminFeedStats
	for _, row := range rows {
		// create a new feed
		var feed Feed
		totalRecords = int(row.TotalCount)
		feed.ID = row.ID
		feed.CreatedAt = row.CreatedAt
		feed.UpdatedAt = row.UpdatedAt
		feed.Name = row.Name
		feed.Url = row.Url
		feed.Version = row.Version
		feed.UserID = row.UserID
		feed.ImgURL = row.ImgUrl
		feed.FeedType = row.FeedType
		feed.FeedDescription = row.FeedDescription
		feed.Is_Hidden = row.IsHidden
		// create a new admin feed
		adminFeed := &AdminFeed{
			Feed:            feed,
			AdminFeedUser:   AdminFeedUser{UserID: row.UserID, Name: row.UserName, Email: row.UserEmail, User_Img: row.UserImg},
			Approval_Status: row.ApprovalStatus,
			Priority:        row.Priority.String,
		}
		// follow count
		adminFeed.FollowCount = row.FollowCount
		// stats
		adminFeedStats.AdminPendingFeedStats.TotalPendingFeeds = row.TotalPendingFeeds.Int64
		adminFeedStats.AdminPendingFeedStats.TotalApprovedFeeds = row.TotalApprovedFeeds.Int64
		adminFeedStats.AdminPendingFeedStats.TotalRejectedFeeds = row.TotalRejectedFeeds.Int64
		adminFeedStats.TotalHiddenFeeds = row.TotalHiddenFeeds.Int64
		adminFeedStats.MostCommonFeedType = row.MostCommonFeedType.String
		// append to the list
		adminFeeds = append(adminFeeds, adminFeed)
	}
	// calculate the metadata
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return adminFeeds, metadata, adminFeedStats, nil
}

func (m AdminModel) AdminUpdateFeed(adminFeed *AdminFeed) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// update the feed
	feed := adminFeed.Feed
	row, err := m.DB.AdminUpdateFeed(ctx, database.AdminUpdateFeedParams{
		ID:              feed.ID,
		UserID:          feed.UserID,
		Name:            feed.Name,
		Url:             feed.Url,
		ImgUrl:          feed.ImgURL,
		FeedType:        feed.FeedType,
		FeedDescription: feed.FeedDescription,
		IsHidden:        feed.Is_Hidden,
		Version:         feed.Version,
		ApprovalStatus:  adminFeed.Approval_Status,
		Priority:        adminFeed.Priority,
	})
	// check for an error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	// update the feed version & updated
	adminFeed.Feed.Version = row.Version
	adminFeed.Feed.UpdatedAt = row.UpdatedAt
	// clean. No error,
	return nil
}
