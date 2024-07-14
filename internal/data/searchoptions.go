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
