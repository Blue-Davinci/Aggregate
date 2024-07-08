package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

var (
	ErrDuplicateFeed   = errors.New("duplicate feed")
	ErrDuplicateFollow = errors.New("duplicate follow")
)

type FeedModel struct {
	DB *database.Queries
}
type Feed struct {
	ID              uuid.UUID `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Name            string    `json:"name"`
	Url             string    `json:"url"`
	Version         int32     `json:"version"`
	UserID          int64     `json:"user_id"`
	ImgURL          string    `json:"img_url"`
	FeedType        string    `json:"feed_type"`
	FeedDescription string    `json:"feed_description"`
}
type FeedFollow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	FeedID    uuid.UUID `json:"feed_id"`
	UserID    int64     `json:"user_id"`
}

func ValidateFeed(v *validator.Validator, feed *Feed) {
	//feed name
	v.Check(feed.Name != "", "name", "must be provided")
	v.Check(len(feed.Name) <= 500, "name", "must not be more than 500 bytes long")
	// feed url
	v.Check(feed.Url != "", "url", "must be provided")
	v.Check(validateUrl(feed.Url), "url", "must be a valid URL")
	// feed image
	v.Check(feed.ImgURL != "", "Image", "url must be provided")
	v.Check(validateUrl(feed.ImgURL), "Image", "must have a valid URL")
	// feed type
	v.Check(feed.FeedType != "", "feed type", "must be provided")
	v.Check(len(feed.FeedType) <= 500, "feed type", "must not be more than 500 bytes long")
	// feed description
	v.Check(feed.FeedDescription != "", "feed description", "must be provided")
	v.Check(len(feed.FeedDescription) <= 500, "feed description", "must not be more than 500 bytes long")
}

func ValidateFeedFollow(v *validator.Validator, feedfollow *FeedFollow) {
	_, isvalid := ValidateUUID(feedfollow.ID.String())
	v.Check(feedfollow.ID != uuid.Nil, "feed id", "must be provided")
	v.Check(isvalid, "feed id", "must be a valid UUID")
}

func (m FeedModel) Insert(feed *Feed) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// version will default to 1
	queryresult, err := m.DB.CreateFeed(ctx, database.CreateFeedParams{
		ID:              uuid.New(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Name:            feed.Name,
		Url:             feed.Url,
		UserID:          feed.UserID,
		ImgUrl:          feed.ImgURL,
		FeedType:        feed.FeedType,
		FeedDescription: feed.FeedDescription,
	})
	// check for an error

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "feeds_url_key"`:
			return ErrDuplicateFeed
		default:
			return err
		}
	}
	// set our details into the feed struct
	feed.ID = queryresult.ID
	feed.CreatedAt = queryresult.CreatedAt
	feed.UpdatedAt = queryresult.UpdatedAt
	feed.Version = queryresult.Version
	feed.UserID = queryresult.UserID

	//fmt.Printf(">> Added a Feed With:\nID: %v\nUser ID: %d", feed.ID, feed.UserID)
	// Return the error if any
	return err
}

func (m FeedModel) GetAllFeeds(name string, url string, filters Filters) ([]*Feed, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.GetAllFeeds(ctx, database.GetAllFeedsParams{
		Column1: name,
		Column2: sql.NullString{String: url, Valid: true}, // Convert string to sql.NullString
		Limit:   int32(filters.limit()),
		Offset:  int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, err
	}
	//fmt.Println("Rows: ", rows)
	totalRecords := 0
	feeds := []*Feed{}
	for _, row := range rows {
		var feed Feed
		totalRecords = int(row.Count)
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
		feeds = append(feeds, &feed)
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	//fmt.Println("Total Records: ", totalRecords)
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return feeds, metadata, nil
}
func (m FeedModel) GetAllFeedsFollowedByUser(userID int64, filters Filters) ([]*FeedFollow, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.GetAllFeedsFollowedByUser(ctx, database.GetAllFeedsFollowedByUserParams{
		UserID: userID,
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, err
	}
	totalRecords := 0
	feedfollows := []*FeedFollow{}
	for _, row := range rows {
		var feedfollow FeedFollow
		totalRecords = int(row.Count)
		feedfollow.ID = row.ID
		feedfollow.CreatedAt = row.CreatedAt
		feedfollow.UpdatedAt = row.UpdatedAt
		feedfollow.FeedID = row.FeedID
		feedfollow.UserID = row.UserID
		feedfollows = append(feedfollows, &feedfollow)
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return feedfollows, metadata, nil
}

func (m FeedModel) CreateFeedFollow(feedfollow *FeedFollow) (*FeedFollow, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// create our feed follow
	queryresult, err := m.DB.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feedfollow.ID,
		UserID:    feedfollow.UserID,
	})

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "feed_follows_user_id_feed_id_key"`:
			return nil, ErrDuplicateFollow
		default:
			return nil, err
		}
	}
	// fill in additional information into our feed
	feedfollow.ID = queryresult.FeedID
	feedfollow.CreatedAt = queryresult.CreatedAt
	feedfollow.UpdatedAt = queryresult.UpdatedAt
	//no logical error, thus return.
	return feedfollow, nil
}

// The DeleteFeedFollow() method accepts the feed follow struct
// and deletes the feed follow record from the database.
func (m FeedModel) DeleteFeedFollow(feedFollow *FeedFollow) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// delete the feed follow
	err := m.DB.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		ID:     feedFollow.ID,
		UserID: feedFollow.UserID,
	})
	fmt.Println("Deleting Feed Follow: ", feedFollow.ID, " || User ID:", feedFollow.UserID)
	// TODO: SQLC - Find a way to check for an already deleted follow without running
	// The delete operation again and counting amount of records returned
	// Currently it still works and passes the "unfollowed" response successfully
	// Idea:
	// Maybe check it from the frontend, where they use a store/cache of the
	// unfollowed feed, and get the result from the metadata, also cached
	// on similar requests, check if feed id is the same, and check if
	// the returned record total is the same as the previous one and raise
	// the already unfollowed error.
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrFeedFollowNotFound
		default:
			return err
		}
	}
	return nil
}

// The urlVerifier() helper function accepts a URL as a string and returns a boolean
// based on whether the URL is valid or not.
func validateUrl(urlstr string) bool {
	u, err := url.Parse(urlstr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// The validateUUID() helper function accepts a UUID and returns a boolean based on
// whether the UUID is valid or not.
func ValidateUUID(feedID string) (uuid.UUID, bool) {
	if feedID == "" {
		return uuid.Nil, false
	}
	parsedfeedID, err := uuid.Parse(feedID)
	return parsedfeedID, err == nil
}
