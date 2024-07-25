package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

var (
	ErrDuplicateFeed   = errors.New("duplicate feed")
	ErrDuplicateFollow = errors.New("duplicate follow")
	ErrConvertingTime  = errors.New("error converting avgTimeBetweenFeeds to sql.NullFloat64")
)

type FeedModel struct {
	DB *database.Queries
}

// This struct will represent the Top fields and contains a Feed struct
// and a Follow_Count field which will be used to represent the number of
// followers a feed has.
type TopFeeds struct {
	Feed         Feed  `json:"feed"`
	Follow_Count int64 `json:"follow_count"`
}

// This struct will represent the Top Creators and contains a User struct
// and the number of feeds created by the user and the total number of
// followers, for each created feed, that the user has.
type TopCreators struct {
	User                       User    `json:"user"`
	Total_Follows              int64   `json:"total_follows"`
	Total_Likes                int64   `json:"total_likes"`
	Total_Created_Feeds        int64   `json:"created_feeds"`
	Average_Time_Between_Feeds float64 `json:"-"`
	Creator_Score              float64 `json:"creator_score"`
}

// This struct will unify the feeds returned providing space for the
// IsFollowed member that will be used to show whether a user follows
// A feed or not. This was necessary so as to get away from frontend
// tabulation of feed follows and feeds setting hte isfollows dynamically
// which would bring a big issue when scaled and data in the 1000s
type FeedsWithFollows struct {
	Feed       Feed      `json:"feed"`
	Follow_ID  uuid.UUID `json:"follow_id"`
	IsFollowed bool      `json:"is_followed"`
}

// This struct will be used to represent the feeds created by a user
// and will be used to return the feeds created by a user and the number
// of followers each feed has.
type FeedsCreatedByUser struct {
	Feed         Feed  `json:"feed"`
	Follow_Count int64 `json:"follow_count"`
}

// This struct will be used when returning detailed information fo a specific
// feed. we also include the creator's information as well as the number of
// liked posts for a specific feed.
type FeedWithStatsInfo struct {
	Feed        Feed  `json:"feed"`
	User        User  `json:"user"`
	Liked_Count int64 `json:"liked_count"`
}

// The Feed struct Represents how our feed struct looks like and is the
// primary model for the feed data.
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
	Is_Hidden       bool      `json:"is_hidden"`
}

// This structs holds information on which feed is followed by which user
// and is used to create a follow record in the database.
type FeedFollow struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	FeedID    uuid.UUID `json:"feed_id"`
	UserID    int64     `json:"-"`
}

// This struct will return the list of feeds followed by a user
type FollowedUserFeeds struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	FeedType  string    `json:"feed_type"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ImgURL    string    `json:"img_url"`
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
	// is hidden
	v.Check(reflect.TypeOf(feed.Is_Hidden).Kind() == reflect.Bool, "is hidden", "must provide a valid type")
}

func ValidateFeedFollow(v *validator.Validator, feedfollow *FeedFollow) {
	_, isvalid := ValidateUUID(feedfollow.ID.String())
	v.Check(feedfollow.ID != uuid.Nil, "feed id", "must be provided")
	v.Check(isvalid, "feed id", "must be a valid UUID")
}

func (m FeedModel) GetFeedWithStats(feedID uuid.UUID) (*FeedWithStatsInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve the feed's information
	feed, err := m.GetFeedByID(feedID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// retrieve the user's and like count information
	row, err := m.DB.GetFeedUserAndStatisticsByID(ctx, feedID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// create our FeedWithStatsInfo struct
	var feedWithStats FeedWithStatsInfo
	feedWithStats.Feed = *feed
	feedWithStats.User.Name = row.UserName
	feedWithStats.User.User_Img = row.UserImgUrl
	feedWithStats.Liked_Count = row.LikedPostsCount
	// we're good, lets return the feed with stats
	return &feedWithStats, nil
}

// The GetFeedByID() method accepts a UUID and returns a pointer to a Feed struct
func (m FeedModel) GetFeedByID(feedID uuid.UUID) (*Feed, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	row, err := m.DB.GetFeedById(ctx, feedID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
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
	return &feed, nil
}

func (m FeedModel) UpdateFeed(feed *Feed) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// update the feed
	err := m.DB.UpdateFeed(ctx, database.UpdateFeedParams{
		ID:              feed.ID,
		Name:            feed.Name,
		Url:             feed.Url,
		ImgUrl:          feed.ImgURL,
		FeedType:        feed.FeedType,
		FeedDescription: feed.FeedDescription,
		IsHidden:        feed.Is_Hidden,
		Version:         feed.Version,
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
	// clean. No error,
	return nil
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
		IsHidden:        feed.Is_Hidden,
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
	return nil
}

func (m FeedModel) GetAllFeeds(name string, url string, filters Filters) ([]*FeedsWithFollows, Metadata, error) {
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
	feeds := []*FeedsWithFollows{}
	for _, row := range rows {
		var feedWithFollow FeedsWithFollows
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
		feed.Is_Hidden = row.IsHidden
		// combine the data
		// set to false by default since this is a general route and the u
		feedWithFollow.IsFollowed = false
		// also for clarity, set this to a nil UUID
		feedWithFollow.Follow_ID = uuid.Nil
		feedWithFollow.Feed = feed

		feeds = append(feeds, &feedWithFollow)
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	//fmt.Println("Total Records: ", totalRecords)
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return feeds, metadata, nil
}

// GetAllFeedsFollowedByUser() returns all feeds with an 'isFollowed' feed that tells the frontend
// whether a feed is followed or not. It also takes in a search string: 'name' if available and searches
// for a feed matching that, i found returns the items as well. We limit by a default of 30 no matter
// whether something is being searched or there is no query.
func (m FeedModel) GetAllFeedsFollowedByUser(userID int64, name string, filters Filters) ([]*FeedsWithFollows, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.GetAllFeedsFollowedByUser(ctx, database.GetAllFeedsFollowedByUserParams{
		UserID:         userID,
		PlaintoTsquery: name,
		Limit:          int32(filters.limit()),
		Offset:         int32(filters.offset()),
	})
	//fmt.Println("Filters: ", filters)
	//check for an error
	if err != nil {
		return nil, Metadata{}, err
	}
	totalRecords := 0
	feedWithFollows := []*FeedsWithFollows{}
	for _, row := range rows {
		var feedWithFollow FeedsWithFollows
		var feedfollow Feed
		totalRecords = int(row.FollowCount)
		feedfollow.ID = row.ID
		feedfollow.CreatedAt = row.CreatedAt
		feedfollow.UpdatedAt = row.UpdatedAt
		feedfollow.Name = row.Name
		feedfollow.Url = row.Url
		feedfollow.Version = row.Version
		feedfollow.UserID = row.UserID
		feedfollow.ImgURL = row.ImgUrl
		feedfollow.FeedType = row.FeedType
		feedfollow.FeedDescription = row.FeedDescription
		feedfollow.Is_Hidden = row.IsHidden
		// combine the data
		feedWithFollow.Feed = feedfollow
		// we set the UUID as a user will need this to unfollow a feed
		feedWithFollow.Follow_ID = row.FollowID
		feedWithFollow.IsFollowed = row.IsFollowed

		feedWithFollows = append(feedWithFollows, &feedWithFollow)
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return feedWithFollows, metadata, nil
}

func (m FeedModel) GetAllFeedsCreatedByUser(userID int64, name string, filters Filters) ([]*FeedsCreatedByUser, Metadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.GetFeedsCreatedByUser(ctx, database.GetFeedsCreatedByUserParams{
		UserID:         userID,
		PlaintoTsquery: name,
		Limit:          int32(filters.limit()),
		Offset:         int32(filters.offset()),
	})
	if err != nil {
		return nil, Metadata{}, err
	}
	totalRecords := 0
	feedCreatedByUsers := []*FeedsCreatedByUser{}
	for _, row := range rows {
		var createdFeed FeedsCreatedByUser
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
		// combine the data
		createdFeed.Feed = feed
		createdFeed.Follow_Count = row.FollowCount

		feedCreatedByUsers = append(feedCreatedByUsers, &createdFeed)
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return feedCreatedByUsers, metadata, nil
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
	feedfollow.ID = queryresult.ID
	feedfollow.CreatedAt = queryresult.CreatedAt
	feedfollow.UpdatedAt = queryresult.UpdatedAt
	feedfollow.UserID = queryresult.UserID
	feedfollow.FeedID = queryresult.FeedID
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

// The GetListOfFollowedFeeds() method returns a list of feeds followed by a user directly
// from the database. It also returns a metadata struct that contains the total records
// and pagination parameters. This route supportd pagination and search via the feed's 'name' parameter.
func (m FeedModel) GetListOfFollowedFeeds(userID int64, name string, filters Filters) ([]*FollowedUserFeeds, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.GetListOfFollowedFeeds(ctx, database.GetListOfFollowedFeedsParams{
		UserID:         userID,
		PlaintoTsquery: name,
		Limit:          int32(filters.limit()),
		Offset:         int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, err
	}
	totalRecords := 0
	followedFeeds := []*FollowedUserFeeds{}
	for _, row := range rows {
		var followedFeed FollowedUserFeeds
		totalRecords = int(row.TotalCount)
		followedFeed.ID = row.ID
		followedFeed.Name = row.Name
		followedFeed.Url = row.Url
		followedFeed.FeedType = row.FeedType
		followedFeed.CreatedAt = row.CreatedAt
		followedFeed.UpdatedAt = row.UpdatedAt
		followedFeed.ImgURL = row.ImgUrl
		followedFeeds = append(followedFeeds, &followedFeed)
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return followedFeeds, metadata, nil
}

// GetTopFeedCreators() returns the top feed creators based on the number of followers
// This route allows for paginations and allows the users to also send a custom length
// or rather pagesize for how many top individuals they need, the default is 5.
func (m FeedModel) GetTopFeedCreators(filters Filters) ([]*TopCreators, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.GetTopFeedCreators(ctx, database.GetTopFeedCreatorsParams{
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, err
	}
	topCreators := []*TopCreators{}
	for _, row := range rows {
		var topCreator TopCreators
		// attach the user to the topcreator struct
		topCreator.User = User{
			Name:     row.Name.String,
			User_Img: row.UserImg.String,
		}
		topCreator.Total_Follows = row.TotalFollows
		topCreator.Total_Likes = row.TotalLikes.Int64
		topCreator.Total_Created_Feeds = row.TotalCreatedFeeds.Int64
		avgTimeBetweenFeeds, ok := row.AvgTimeBetweenFeeds.(float64)
		if !ok {
			return nil, ErrConvertingTime
		}
		topCreator.Average_Time_Between_Feeds = avgTimeBetweenFeeds
		// calculate the creator score
		topCreator.Creator_Score = m.scoreCalculationAlgorithm(&topCreator)
		// append the topcreator to the topcreators slice
		topCreators = append(topCreators, &topCreator)
	}
	return topCreators, nil
}

func (m FeedModel) GetTopFollowedFeeds(filters Filters) ([]*TopFeeds, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// retrieve our data
	rows, err := m.DB.GetTopFollowedFeeds(ctx, int32(filters.limit()))
	//check for an error
	if err != nil {
		return nil, err
	}
	//fmt.Println("Rows: ", rows)
	topFeeds := []*TopFeeds{}
	for _, row := range rows {
		var topfeed TopFeeds
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
		// attach the feed to the topfeed struct
		topfeed.Feed = feed
		topfeed.Follow_Count = row.FollowCount
		topFeeds = append(topFeeds, &topfeed)
	}
	return topFeeds, nil
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
