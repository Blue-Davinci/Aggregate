package data

import (
	"context"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/google/uuid"
)

type SearchOptionsDataModel struct {
	DB *database.Queries
}

// struct to hold a struct returned for all available distinct feed
type SearchOptionFeedDetail struct {
	Feed_ID   uuid.UUID `json:"feed_id"`
	Feed_Name string    `json:"feed_name"`
}

type SearchOptionFeedType struct {
	Feed_ID   int    `json:"feed_id"`
	Feed_Type string `json:"feed_type"`
}

// The ID's will be used for interopolations for the frontend

// GetFeedSearchOptions() returns all available distinct feeds
// we have in the database. For the ID's, we will use their UUID's
func (m SearchOptionsDataModel) GetFeedSearchOptions() ([]*SearchOptionFeedDetail, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	feeds, err := m.DB.GetFeedSearchOptions(ctx)
	if err != nil {
		return nil, err
	}
	feedDetails := []*SearchOptionFeedDetail{}
	for _, feed := range feeds {
		var feedDetail SearchOptionFeedDetail
		feedDetail.Feed_ID = feed.ID
		feedDetail.Feed_Name = feed.Name
		feedDetails = append(feedDetails, &feedDetail)
	}
	return feedDetails, nil
}

// GetFeedTypeSearchOptions() returns all available distinct feed types
// we have in the database. For the ID's, we will just use the indexes
func (m SearchOptionsDataModel) GetFeedTypeSearchOptions() ([]*SearchOptionFeedType, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	feeds, err := m.DB.GetFeedTypeSearchOptions(ctx)
	if err != nil {
		return nil, err
	}
	feedTypes := []*SearchOptionFeedType{}
	for i, feed_type := range feeds {
		var feedType SearchOptionFeedType
		feedType.Feed_ID = i + 1
		feedType.Feed_Type = feed_type
		feedTypes = append(feedTypes, &feedType)
	}
	return feedTypes, nil
}
