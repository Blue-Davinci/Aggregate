package data

import (
	"context"
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

type AdminUser struct {
	ID          int64       `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	Name        string      `json:"name"`
	Email       string      `json:"email"`
	Password    string      `json:"-"` // we omit this yk, Eve for godmode we
	Activated   bool        `json:"activated"`
	Version     int         `json:"version"`
	User_Img    string      `json:"user_img"`
	Permissions Permissions `json:"permissions"`
}

type AdminStatistics struct {
	UserStatistics         UserStatistics         `json:"user_statistics"`
	SubscriptionStatistics SubscriptionStatistics `json:"subscription_statistics"`
	CommentStatistics      CommentStatistics      `json:"comment_statistics"`
}

type UserStatistics struct {
	Total_Users   int64   `json:"total_users"`
	Active_Users  int64   `json:"active_users"`
	New_Signups   int64   `json:"new_signups"`
	Total_Revenue float64 `json:"total_revenue"`
}

type SubscriptionStatistics struct {
	Total_Revenue            float64 `json:"total_revenue"`
	Active_Subscriptions     int64   `json:"active_subscriptions"`
	Cancelled_Subscriptions  int64   `json:"cancelled_subscriptions"`
	Expired_Subscriptions    int64   `json:"expired_subscriptions"`
	Most_Used_Payment_Method string  `json:"most_used_payment_method"`
}

type CommentStatistics struct {
	Total_Comments  int64 `json:"total_comments"`
	Recent_Comments int64 `json:"recent_comments"`
}

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
