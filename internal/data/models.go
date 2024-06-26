package data

import (
	"errors"

	"github.com/blue-davinci/aggregate/internal/database"
)

var (
	ErrRecordNotFound     = errors.New("feeds record not found")
	ErrFeedFollowNotFound = errors.New("feed already unfollowed")
	ErrEditConflict       = errors.New("edit conflict")
)

// Holds our models. Makes it easy for dependancy injection for each app instance
type Models struct {
	Users       UserModel
	ApiKey      ApiKeyModel
	Feeds       FeedModel
	RSSFeedData RSSFeedDataModel
	//feed models
}

// Returns a new model instance
func NewModels(db *database.Queries) Models {
	return Models{
		Users:       UserModel{DB: db},
		ApiKey:      ApiKeyModel{DB: db},
		Feeds:       FeedModel{DB: db},
		RSSFeedData: RSSFeedDataModel{DB: db},
	}
}
