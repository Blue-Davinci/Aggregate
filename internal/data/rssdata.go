package data

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/microcosm-cc/bluemonday"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/atom"
)

// Constants for our RSS Feed Scraper that won't be set using flags
// ResponseContextTimeout - This is the timeout for our context when fetching feeds
const (
	// Context timeout
	ResponseContextTimeout = 30 * time.Second
	// Default Image Url
	//https://images.pexels.com/photos/17300603/pexels-photo-17300603/free-photo-of-a-little-girl-running-in-a-park-next-to-a-man-making-a-large-soap-bubble.jpeg?auto=compress&cs=tinysrgb&w=1260&h=750&dpr=1
	DefaultImageURL = "https://images.unsplash.com/photo-1542396601-dca920ea2807?q=80&w=1351&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D"
)

var (
	// Context Error
	ErrContextDeadline        = errors.New("timeout exceeded while fetching feeds")
	ErrUnableToDetectFeedType = errors.New("unable to detect the feed type in the url")
	ErrDuplicateFavorite      = errors.New("duplicate favorite")
	ErrPostNotFound           = errors.New("post not found")
)

// RSSFeedDataModel is a struct that represents what our Post looks like
type RSSFeedDataModel struct {
	DB *database.Queries
}

// This is our main struct that is returned from our post endpoint and returns
// posts with an isFavorite field
type RSSFeedWithFavorite struct {
	RSSFeed    *RSSFeed `json:"feed"`
	IsFavorite bool     `json:"isFavorite"`
	IsFollowed bool     `json:"isFollowed"`
}

// RSSFeed is a struct that represents what our RSS Feed looks like
type RSSFeed struct {
	ID        uuid.UUID `json:"id"`
	Createdat time.Time `json:"created_at"`
	Updatedat time.Time `json:"updated_at"`
	Feed_ID   uuid.UUID `json:"feed_id"`
	Channel   struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
	RetryMax   int32 `json:"-"`
	StatusCode int32 `json:"-"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Content     string `xml:"content"`
	PubDate     string `xml:"pubDate"`
	ImageURL    string `xml:"image_url"`
}

// We make a solo struct that will hold a returned Post Favorite
type RSSPostFavorite struct {
	ID         int64     `json:"id"`
	Post_ID    uuid.UUID `json:"post_id"`
	Feed_ID    uuid.UUID `json:"feed_id"`
	User_ID    int64     `json:"-"`
	Created_At time.Time `json:"created_at"`
}

func ValidateFavoritePost(v *validator.Validator, favoritePost *RSSPostFavorite) {
	_, isvalidPostID := ValidateUUID(favoritePost.Post_ID.String())
	_, isvalidFeedID := ValidateUUID(favoritePost.Feed_ID.String())
	v.Check(isvalidPostID, "feed id", "must be a valid UUID")
	v.Check(isvalidFeedID, "feed id", "must be a valid UUID")
}

func ValidatePostID(v *validator.Validator, favoritePost *RSSPostFavorite) {
	_, isvalidPostID := ValidateUUID(favoritePost.Post_ID.String())
	v.Check(isvalidPostID, "post id", "must be a valid UUID")
}

// GetNextFeedsToFetch() will get the next feeds to fetch for our scraper after which
// we return an error, if any, and the feeds we found
func (m RSSFeedDataModel) GetNextFeedsToFetch(noofroutines int, feedRequestInterval int) ([]database.Feed, error) {
	// This will get the next feeds to fetch
	feeds, err := m.DB.GetNextFeedsToFetch(context.Background(), int32(noofroutines))
	if err != nil {
		return nil, err
	}
	return feeds, nil
}

// MarkFeedAsFetched() will mark the feed as fetched by updating the last_fetched field
func (m RSSFeedDataModel) MarkFeedAsFetched(feed uuid.UUID) error {
	_, err := m.DB.MarkFeedAsFetched(context.Background(), feed)
	if err != nil {
		return err
	}
	return nil
}

// GetFollowedRssPostsForUser() is our main RSS Posts function that serves as both
// the endpoint fir getting all posts and also for getting all posts filtered by the UUID
// or searched by the itemtitle/post title
// We return this as a slice of RSSFeed structs but with an isfavorite field
// to show whether the post is in the user's favorites so that the frontend can set it
// as a favorite or not
func (m RSSFeedDataModel) GetFollowedRssPostsForUser(userID int64, feed_name string, feed_id uuid.UUID, filters Filters) ([]*RSSFeedWithFavorite, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rssFeedPosts, err := m.DB.GetFollowedRssPostsForUser(ctx, database.GetFollowedRssPostsForUserParams{
		UserID:  userID,
		Column2: feed_name,
		Column3: feed_id,
		Limit:   int32(filters.limit()),
		Offset:  int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, err
	}
	//make a var for our metadata
	var metadata Metadata
	totalRecords := 0
	// make a store for our processed feeds
	rssFeedWithFavorites := []*RSSFeedWithFavorite{}
	for _, row := range rssFeedPosts {
		// create a singly rss feed with favorite
		var rssFeedWithFavorite RSSFeedWithFavorite
		// create a single rss feed
		var rssFeed RSSFeed
		totalRecords = int(row.TotalCount)
		// General infor
		rssFeed.ID = row.ID
		rssFeed.Createdat = row.CreatedAt
		rssFeed.Updatedat = row.UpdatedAt
		rssFeed.Feed_ID = row.FeedID
		// Channel info
		rssFeed.Channel.Title = row.Channeltitle
		rssFeed.Channel.Description = row.Channeldescription.String
		rssFeed.Channel.Link = row.Channelurl.String
		rssFeed.Channel.Language = row.Channellanguage.String
		// Item Info
		rssFeed.Channel.Item = append(rssFeed.Channel.Item, RSSItem{
			Title:       row.Itemtitle,
			Link:        row.Itemurl,
			Description: row.Itemdescription.String,
			Content:     row.Itemcontent.String,
			PubDate:     row.ItempublishedAt.String(),
			ImageURL:    row.ImgUrl,
		})
		// aggregate our data to the final struct
		rssFeedWithFavorite.RSSFeed = &rssFeed
		rssFeedWithFavorite.IsFavorite = row.IsFavorite
		rssFeedWithFavorite.IsFollowed = true
		//append our feed to the final slice
		metadata = calculateMetadata(totalRecords, filters.Page, filters.PageSize)
		rssFeedWithFavorites = append(rssFeedWithFavorites, &rssFeedWithFavorite)
	}

	return rssFeedWithFavorites, metadata, nil
}

// We need to add isFollowe to the return item so that a user can know if it was
// followed or not incase a user clicks on a shared item
func (m RSSFeedDataModel) GetRSSFeedByID(userID int64, feedID uuid.UUID) (*RSSFeedWithFavorite, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// get our feed
	feed, err := m.DB.GetRssPostByPostID(ctx, database.GetRssPostByPostIDParams{
		UserID:  userID,
		Column2: feedID,
	})
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrPostNotFound
		default:
			return nil, err
		}
	}
	// create a singly rss feed with favorite
	var rssFeedWithFavorite RSSFeedWithFavorite
	// create a single rss feed
	var rssFeed RSSFeed
	// General infor
	rssFeed.ID = feed.ID
	rssFeed.Createdat = feed.CreatedAt
	rssFeed.Updatedat = feed.UpdatedAt
	rssFeed.Feed_ID = feed.FeedID
	// Channel info
	rssFeed.Channel.Title = feed.Channeltitle
	rssFeed.Channel.Description = feed.Channeldescription.String
	rssFeed.Channel.Link = feed.Channelurl.String
	rssFeed.Channel.Language = feed.Channellanguage.String
	// Item Info
	rssFeed.Channel.Item = append(rssFeed.Channel.Item, RSSItem{
		Title:       feed.Itemtitle,
		Link:        feed.Itemurl,
		Description: feed.Itemdescription.String,
		Content:     feed.Itemcontent.String,
		PubDate:     feed.ItempublishedAt.String(),
		ImageURL:    feed.ImgUrl,
	})
	// aggregate our data to the final struct
	// we ad an isFollowed on this one so that if a user gets a url and has sees the post
	// this will allow the frontend to know if it should give the user the option to follow the feed
	// or rather suggest to the user to follow the feed responsible for the post.
	rssFeedWithFavorite.RSSFeed = &rssFeed
	rssFeedWithFavorite.IsFavorite = feed.IsFavorite.(bool)
	rssFeedWithFavorite.IsFollowed = feed.IsFollowedFeed.(bool)
	return &rssFeedWithFavorite, nil
}

// CreateRssFeed() Is a scraper hooked function which will recieve all data scrapped
// and will proceed to save it in the database.
func (m RSSFeedDataModel) CreateRssFeedPost(rssFeed *RSSFeed, feedID *uuid.UUID) error {
	// Get channel Info
	ChannelTitle := rssFeed.Channel.Title
	ChannelUrl := rssFeed.Channel.Link
	ChannelDescription := rssFeed.Channel.Description
	ChannelLanguage := rssFeed.Channel.Language
	for _, item := range rssFeed.Channel.Item {
		// We use dateparse to parse a variety of possible date/time data rather than using
		// the time.Parse() function which is more strict.
		// We use ParseAny()
		publishedAt, err := dateparse.ParseAny(item.PubDate)
		if err != nil {
			continue
		}
		_, err = m.DB.CreateRssFeedPost(context.Background(), database.CreateRssFeedPostParams{
			// Default Info
			ID:        uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			// Channel info
			Channeltitle:       ChannelTitle,
			Channelurl:         sql.NullString{String: ChannelUrl, Valid: ChannelUrl != ""},
			Channeldescription: sql.NullString{String: ChannelDescription, Valid: ChannelDescription != ""},
			Channellanguage:    sql.NullString{String: ChannelLanguage, Valid: ChannelLanguage != ""},
			// Item Info
			Itemtitle:       item.Title,
			Itemdescription: sql.NullString{String: item.Description, Valid: rssFeed.Channel.Description != ""},
			Itemcontent:     sql.NullString{String: item.Content, Valid: item.Content != ""},
			ItempublishedAt: publishedAt,
			Itemurl:         item.Link,
			ImgUrl:          item.ImageURL,
			FeedID:          *feedID,
		})
		// Our db should not contain the same  URL/Post twice, so we just ignore this error (is it an error really?)
		// and actually print real ones.
		if err != nil && err.Error() != `pq: duplicate key value violates unique constraint "rssfeed_posts_itemurl_key"` {
			fmt.Println("Couldn't create post for: ", item.Title, "Error: ", err)
		}
	}
	return nil
}

// GetRSSFavoritePostsForUser() returns the RSS Posts that a user has favorited
// It will take in the userID and return a slice of RSSPostFavorite structs and an error if any
// Should be used in tandem with GetFollowedRssPostsForUser()
func (m RSSFeedDataModel) GetRSSFavoritePostsForUser(userID int64) ([]*RSSPostFavorite, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Get favorited posts, passing in the user's ID as the parameter
	rssFeedFavoritePosts, err := m.DB.GetRSSFavoritePostsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	// make a store for our processed feeds
	rssPostsFavorites := []*RSSPostFavorite{}
	for _, row := range rssFeedFavoritePosts {
		var rssPostFavorite RSSPostFavorite
		// General infor
		rssPostFavorite.ID = row.ID
		rssPostFavorite.Post_ID = row.PostID
		rssPostFavorite.Feed_ID = row.FeedID
		rssPostFavorite.User_ID = row.UserID
		rssPostFavorite.Created_At = row.CreatedAt
		//append our feed to the final slice
		rssPostsFavorites = append(rssPostsFavorites, &rssPostFavorite)
	}
	return rssPostsFavorites, nil
}

// CreateRSSFavoritePost() will create a new RSS Favorite Post for a user
//
//	This method recieves the user's ID and the post ID and keeps track of which
//	posts the user has favorited
func (m RSSFeedDataModel) CreateRSSFavoritePost(userID int64, rssFavoritePost *RSSPostFavorite) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Create a new RSS Favorite Post
	queryresult, err := m.DB.CreateRSSFavoritePost(ctx, database.CreateRSSFavoritePostParams{
		PostID:    rssFavoritePost.Post_ID,
		FeedID:    rssFavoritePost.Feed_ID,
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	})
	// postfavorites_post_id_key
	if err != nil {
		switch {
		case err.Error() != `pq: duplicate key value violates unique constraint "rssfeed_posts_itemurl_key"`:
			return ErrDuplicateFavorite
		default:
			return err
		}
	}
	// Set additional fields
	rssFavoritePost.ID = queryresult.ID
	rssFavoritePost.User_ID = queryresult.UserID
	rssFavoritePost.Created_At = queryresult.CreatedAt
	return nil
}

// DeleteRSSFavoritePost() will delete an RSS Favorite Post for a user
// It will accept a user's ID and the post's ID that needs to be removed
// from their favorites
func (m RSSFeedDataModel) DeleteRSSFavoritePost(userID int64, rssFavoritePost *RSSPostFavorite) error {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("Recieved, ID:", userID, "|| Value: ", rssFavoritePost.Post_ID)
	// Delete the RSS Favorite Post
	err := m.DB.DeleteRSSFavoritePost(ctx, database.DeleteRSSFavoritePostParams{
		PostID: rssFavoritePost.Post_ID,
		UserID: userID,
	})
	if err != nil {
		switch {
		case err == sql.ErrNoRows:
			return ErrPostNotFound
		default:
			fmt.Println("The error: ", err)
			return err
		}
	}
	return nil
}

// GetRandomRSSPosts Will return a small subset of information about posts for non-registered users
// It doesn not contain all information, it will only have very little information enough to wet
// their appettite.
func (m RSSFeedDataModel) GetRandomRSSPosts(feedID uuid.UUID, filters Filters) ([]*RSSFeed, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rssFeedPosts, err := m.DB.GetRandomRSSPosts(ctx, database.GetRandomRSSPostsParams{
		FeedID: feedID,
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrPostNotFound
		default:
			return nil, err
		}
	}
	// make a store for our processed feeds
	rssPosts := []*RSSFeed{}
	for _, row := range rssFeedPosts {
		var rssPost RSSFeed
		// General infor
		rssPost.ID = row.ID
		// Channel info
		rssPost.Channel.Title = row.Channeltitle
		// Item Info
		rssPost.Channel.Item = append(rssPost.Channel.Item, RSSItem{
			Title:       row.Itemtitle,
			Description: row.Itemdescription.String,
			ImageURL:    row.ImgUrl,
		})
		//append our feed to the final slice
		rssPosts = append(rssPosts, &rssPost)
	}
	return rssPosts, nil
}

// This will get the RSS Favorite Posts for a user only, it gets the User ID and the filters
// and returns a subset of all posts followed by a user
func (m RSSFeedDataModel) GetRSSFavoritePostsOnlyForUser(userID int64, feed_name string, feed_id uuid.UUID, filters Filters) ([]*RSSFeedWithFavorite, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rssFeedPosts, err := m.DB.GetRSSFavoritePostsOnlyForUser(ctx, database.GetRSSFavoritePostsOnlyForUserParams{
		UserID:  userID,
		Column2: feed_name,
		Column3: feed_id,
		Limit:   int32(filters.limit()),
		Offset:  int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, err
	}
	// make a totals and favoritePosts variables to hold the results
	totalRecords := 0
	favoritePosts := []*RSSFeedWithFavorite{}
	var metadata Metadata
	for _, row := range rssFeedPosts {
		var favoritePost RSSFeedWithFavorite
		var rssPost RSSFeed
		totalRecords = int(row.TotalCount)
		// General infor
		rssPost.ID = row.ID
		rssPost.Createdat = row.CreatedAt
		rssPost.Updatedat = row.UpdatedAt
		rssPost.Feed_ID = row.FeedID
		// Channel info
		rssPost.Channel.Title = row.Channeltitle
		rssPost.Channel.Description = row.Channeldescription.String
		rssPost.Channel.Link = row.Channelurl.String
		rssPost.Channel.Language = row.Channellanguage.String
		// Item Info
		rssPost.Channel.Item = append(rssPost.Channel.Item, RSSItem{
			Title:       row.Itemtitle,
			Link:        row.Itemurl,
			Description: row.Itemdescription.String,
			Content:     row.Itemcontent.String,
			PubDate:     row.ItempublishedAt.String(),
			ImageURL:    row.ImgUrl,
		})
		//Aggregate the data
		favoritePost.RSSFeed = &rssPost
		favoritePost.IsFavorite = row.IsFavorite
		favoritePost.IsFollowed = row.IsFollowedFeed.(bool)
		//append our feed to the final slice
		metadata = calculateMetadata(totalRecords, filters.Page, filters.PageSize)
		favoritePosts = append(favoritePosts, &favoritePost)
	}
	return favoritePosts, metadata, nil
}

// =======================================================================================================================
//
//	SCRAPER
//
// =======================================================================================================================

// GetRSSFeeds() is a method that will fetch the RSS feeds from our RSS Feed URL
// It will take in the retryMax, clientTimeout and the URL of the feed to fetch and
// Use our Decoder method to Decode the XML body recieved from the feed.
// It will return an RSSFeed struct and an error if any
func (m RSSFeedDataModel) GetRSSFeeds(retryMax, clientTimeout int, url string, sanitizer *bluemonday.Policy) (RSSFeed, error) {
	// create a retrayable client with our own settings
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.HTTPClient.Timeout = time.Duration(clientTimeout) * time.Second
	retryClient.Backoff = retryablehttp.LinearJitterBackoff
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryClient.Logger = nil

	// Create a new request with context for timeout
	req, err := retryablehttp.NewRequest("GET", url, nil)
	if err != nil {
		//fmt.Println("++++++>>>>>>>> err: ", err)
		return RSSFeed{}, err
	}
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), ResponseContextTimeout)
	defer cancel() // Ensure the context is cancelled to free resources
	req = req.WithContext(ctx)

	// Perform the request with retries
	resp, err := retryClient.Do(req)
	if err != nil {
		fmt.Println("++++++Client Rec err: ", err)
		switch {
		case strings.Contains(err.Error(), "context deadline exceeded"):
			return RSSFeed{}, ErrContextDeadline
		default:
			return RSSFeed{}, err
		}
	}
	defer resp.Body.Close()
	// Initialize a new RSSFeed struct
	rssFeed := RSSFeed{}
	// Decode the response using RssFeedDecoder() expecting an RSSFeed struct
	err = RssFeedDecoderDecider(url, &rssFeed, sanitizer, resp)
	if err != nil {
		fmt.Println("++++++>Dec Dec err: ", err)
		switch {
		case strings.Contains(err.Error(), "context deadline exceeded"):
			return RSSFeed{}, ErrContextDeadline
		case strings.Contains(err.Error(), "feed type"):
			return RSSFeed{RetryMax: int32(retryMax), StatusCode: int32(resp.StatusCode)}, ErrUnableToDetectFeedType
		default:
			return RSSFeed{}, err
		}
	}

	return rssFeed, nil
}

// RssFeedDecoder() will decide which type of URL we are fetching i.e. Atom or RSS
// and then choose different decoders for each type of feed
func RssFeedDecoderDecider(url string, rssFeed *RSSFeed, sanitizer *bluemonday.Policy, resp *http.Response) error {
	// Read the entire response body into a byte slice
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Attempt to parse using gofeed
	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(data))
	if err == nil && feed != nil {
		convertGofeedToRSSFeed(rssFeed, feed, sanitizer)
		return nil
	} else if err != nil {
		// Log or return specific error for gofeed parsing failure
		return fmt.Errorf("gofeed parsing error: %w", err)
	}

	// Attempt to parse using atom parser
	atomParser := &atom.Parser{}
	atomFeed, err := atomParser.Parse(bytes.NewReader(data))
	if err == nil && atomFeed != nil {
		convertAtomfeedToRSSFeed(rssFeed, atomFeed, sanitizer)
		return nil
	} else if err != nil {
		// Log or return specific error for atom parsing failure
		return fmt.Errorf("atom parsing error: %w", err)
	}

	// If all parsing attempts fail, return a generic error
	return fmt.Errorf("unable to parse feed from URL: %s", url)
}

// =======================================================================================================================
//
//	CONVERTORS
//
// =======================================================================================================================

// convertAtomfeedToRSSFeed() will convert an atom.Feed struct to our RSSFeed struct
// This is done by copying the fields from the atom.Feed struct to our RSSFeed struct
// acknowledging the differences in field items and field entries
func convertAtomfeedToRSSFeed(rssFeed *RSSFeed, feed *atom.Feed, sanitizer *bluemonday.Policy) {
	if rssFeed == nil || feed == nil {
		fmt.Println("RSSFeed pointer or atom.Feed pointer is nil")
		return
	}
	//proceed to fill the main channel fields
	rssFeed.Channel.Title = sanitizer.Sanitize(feed.Title)
	rssFeed.Channel.Description = sanitizer.Sanitize(feed.Subtitle)
	// Grab our first link as the main link for the channel
	if len(feed.Links) > 0 {
		rssFeed.Channel.Link = sanitizer.Sanitize(feed.Links[0].Href)
	}
	rssFeed.Channel.Language = sanitizer.Sanitize(feed.Language)
	// Use the correct field for Atom entries, which is `Entries` instead of `Items` as for RSS feeds
	rssFeed.Channel.Item = make([]RSSItem, len(feed.Entries)) // Allocate space for entries
	for i, entry := range feed.Entries {
		// As like RSS feeds, we use a default image URL if no image is found
		// We also use the link property to search for any image URLs
		imageURL := DefaultImageURL
		for _, link := range entry.Links {
			if link.Rel == "enclosure" || link.Type == "image/jpeg" || link.Type == "image/png" {
				imageURL = link.Href
				break // Found an image URL, exit the loop
			}
		}
		rssFeed.Channel.Item[i] = RSSItem{
			Title:       sanitizer.Sanitize(entry.Title),
			Link:        sanitizer.Sanitize(entry.Links[0].Href),
			Description: sanitizer.Sanitize(entry.Summary),
			Content:     sanitizer.Sanitize(entry.Content.Value),
			PubDate:     sanitizer.Sanitize(entry.Published),
			ImageURL:    imageURL, // sanitizer.Sanitize(imageURL)
		}
	}
}

// convertGofeedToRSSFeed() will convert a gofeed.Feed struct to our RSSFeed struct
// This is done by copying the fields from the gofeed.Feed struct to our RSSFeed struct
func convertGofeedToRSSFeed(rssFeed *RSSFeed, feed *gofeed.Feed, sanitizer *bluemonday.Policy) {
	if rssFeed == nil || feed == nil {
		fmt.Println("RSSFeed pointer or gofeed.Feed pointer is nil")
		return
	}
	// Fill the main channel fields
	rssFeed.Channel.Title = sanitizer.Sanitize(feed.Title)
	rssFeed.Channel.Link = sanitizer.Sanitize(feed.Link)
	rssFeed.Channel.Description = sanitizer.Sanitize(feed.Description)
	rssFeed.Channel.Language = sanitizer.Sanitize(feed.Language)
	// Use the correct field for RSS items
	rssFeed.Channel.Item = make([]RSSItem, len(feed.Items)) // Allocate space for items
	for i, item := range feed.Items {
		// As like Atom feeds, we use a default image URL if no image is found
		imageURL := DefaultImageURL
		if item.Image != nil {
			imageURL = item.Image.URL
		}
		/*
			// save dcontent to file
			fileName := fmt.Sprintf("item_%d.txt", i)
			filePath := filepath.Join("output", fileName)
			err := os.MkdirAll("output", os.ModePerm)
			if err != nil {
				fmt.Printf("Error creating directory: %v\n", err)
				continue
			}
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error creating file: %v\n", err)
				continue
			}
			defer file.Close()
			_, err = file.WriteString(item.Content)
			if err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				continue
			}
			/// -----------------
		*/
		rssFeed.Channel.Item[i] = RSSItem{
			Title:       sanitizer.Sanitize(item.Title),
			Link:        sanitizer.Sanitize(item.Link),
			Description: sanitizer.Sanitize(item.Description),
			Content:     sanitizer.Sanitize(item.Content),
			PubDate:     sanitizer.Sanitize(item.Published),
			ImageURL:    imageURL, // sanitizer.Sanitize(imageURL)
		}
	}
}
