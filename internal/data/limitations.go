package data

import (
	"context"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
)

type LimitationsModel struct {
	DB *database.Queries
}

type LimitationsItmes struct {
	UserID         int64 `json:"user_id"`
	Followed_Feeds int64 `json:"followed_feeds"`
	Created_Feeds  int64 `json:"created_feeds"`
	Comments_Today int64 `json:"comments_today"`
}

// Get the limitations for a user
func (m LimitationsModel) GetUserLimitations(userID int64) (*LimitationsItmes, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Query the database for the limitations
	queryResult, err := m.DB.GetUserLimitations(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Create a new LimitationsItmes struct
	limitations := LimitationsItmes{
		UserID:         queryResult.UserID,
		Followed_Feeds: queryResult.FollowedFeeds,
		Created_Feeds:  queryResult.CreatedFeeds,
		Comments_Today: queryResult.CommentsToday,
	}

	// Return the limitations
	return &limitations, nil
}
