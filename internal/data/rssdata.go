package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/mmcdole/gofeed"
	"github.com/mmcdole/gofeed/atom"
)

// Constants for our RSS Feed Scraper that won't be set using flags
// ResponseContextTimeout - This is the timeout for our context when fetching feeds
const (
	// Context timeout
	ResponseContextTimeout = 30 * time.Second
	// Default Image Url
	DefaultImageURL = "https://rss.com/blog/wp-content/uploads/2019/10/social_style_3_rss-512-1.png"
)

var (
	// Context Error
	ErrContextDeadline        = errors.New("timeout exceeded while fetching feeds")
	ErrUnableToDetectFeedType = errors.New("unable to detect the feed type in the url")
)

type RSSFeedDataModel struct {
	DB *database.Queries
}
type RSSFeed struct {
	ID        uuid.UUID `json:"id"`
	Createdat time.Time `json:"created_at"`
	Updatedat time.Time `json:"updated_at"`
	Channel   struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	ImageURL    string `xml:"image_url"`
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

func (m RSSFeedDataModel) GetFollowedRssPostsForUser(userID int64, filters Filters) ([]*RSSFeed, Metadata, error) {
	// create our timeout context. All of them will just be 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	rssFeedPosts, err := m.DB.GetFollowedRssPostsForUser(ctx, database.GetFollowedRssPostsForUserParams{
		UserID: userID,
		Limit:  int32(filters.limit()),
		Offset: int32(filters.offset()),
	})
	//check for an error
	if err != nil {
		return nil, Metadata{}, err
	}
	//make a var for our metadata
	var metadata Metadata
	totalRecords := 0
	// make a store for our processed feeds
	rssPosts := []*RSSFeed{}
	for _, row := range rssFeedPosts {
		var rssFeed RSSFeed
		totalRecords = int(row.Count)
		// General infor
		rssFeed.ID = row.ID
		rssFeed.Createdat = row.CreatedAt
		rssFeed.Updatedat = row.UpdatedAt
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
			PubDate:     row.ItempublishedAt.String(),
			ImageURL:    row.ImgUrl,
		})
		//append our feed to the final slice
		metadata = calculateMetadata(totalRecords, filters.Page, filters.PageSize)
		rssPosts = append(rssPosts, &rssFeed)
	}

	return rssPosts, metadata, nil
}

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

// =======================================================================================================================
//
//	SCRAPER
//
// =======================================================================================================================

// GetRSSFeeds() is a method that will fetch the RSS feeds from our RSS Feed URL
// It will take in the retryMax, clientTimeout and the URL of the feed to fetch and
// Use our Decoder method to Decode the XML body recieved from the feed.
// It will return an RSSFeed struct and an error if any
func (m RSSFeedDataModel) GetRSSFeeds(retryMax, clientTimeout int, url string) (RSSFeed, error) {
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
	err = RssFeedDecoder(url, &rssFeed, resp)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "context deadline exceeded"):
			return RSSFeed{}, ErrContextDeadline
		case strings.Contains(err.Error(), "feed type"):
			return RSSFeed{}, ErrUnableToDetectFeedType
		default:
			return RSSFeed{}, err
		}
	}
	return rssFeed, nil
}

// RssFeedDecoder() will decide which type of URL we are fetching i.e. Atom or RSS
// and then choose different decoders for each type of feed
func RssFeedDecoder(url string, rssFeed *RSSFeed, resp *http.Response) error {
	// Check if the feed is an atom feed or a normal RSS Feed
	// We try and convert it using GoFeed Parser First
	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		// If we get an error here, lets check if it's a context deadline exceeded error
		// and return it specially. There's no need to continue processing the url, so we will return.
		return err
	}
	if feed == nil {
		// If it's atom, we parse original body to atom feed
		fp := atom.Parser{}
		feed, err := fp.Parse(resp.Body)
		if err != nil {
			//fmt.Println(">>>>>>>{}{}{}{}{}>>>>>>>>>>>>>")
			return err
		}
		// Then we call our function to convert the atom feed to our standard RSS Feed
		convertAtomfeedToRSSFeed(rssFeed, feed)
	} else if feed.FeedType == "rss" {
		// Otherwisse, it's a normal RSS Feed, so we call our function to convert it
		// to our standard RSS Feed
		convertGofeedToRSSFeed(rssFeed, feed)
	}
	return nil
}

// =======================================================================================================================
//
//	CONVERTORS
//
// =======================================================================================================================

// convertAtomfeedToRSSFeed() will convert an atom.Feed struct to our RSSFeed struct
// This is done by copying the fields from the atom.Feed struct to our RSSFeed struct
// acknowledging the differences in field items and field entries
func convertAtomfeedToRSSFeed(rssFeed *RSSFeed, feed *atom.Feed) {
	if rssFeed == nil || feed == nil {
		fmt.Println("RSSFeed pointer or atom.Feed pointer is nil")
		return
	}
	//proceed to fill the main channel fields
	rssFeed.Channel.Title = feed.Title
	rssFeed.Channel.Description = feed.Subtitle
	// Grab our first link as the main link for the channel
	if len(feed.Links) > 0 {
		rssFeed.Channel.Link = feed.Links[0].Href
	}
	rssFeed.Channel.Language = feed.Language
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
			Title:       entry.Title,
			Link:        entry.Links[0].Href, // we'll just pick the first link
			Description: entry.Summary,
			PubDate:     entry.Published,
			ImageURL:    imageURL,
		}
	}
}

// convertGofeedToRSSFeed() will convert a gofeed.Feed struct to our RSSFeed struct
// This is done by copying the fields from the gofeed.Feed struct to our RSSFeed struct
func convertGofeedToRSSFeed(rssFeed *RSSFeed, feed *gofeed.Feed) {
	if rssFeed == nil || feed == nil {
		fmt.Println("RSSFeed pointer or gofeed.Feed pointer is nil")
		return
	}
	// Fill the main channel fields
	rssFeed.Channel.Title = feed.Title
	rssFeed.Channel.Link = feed.Link
	rssFeed.Channel.Description = feed.Description
	rssFeed.Channel.Language = feed.Language
	// Use the correct field for RSS items
	rssFeed.Channel.Item = make([]RSSItem, len(feed.Items)) // Allocate space for items
	for i, item := range feed.Items {
		// As like Atom feeds, we use a default image URL if no image is found
		imageURL := DefaultImageURL
		if item.Image != nil {
			imageURL = item.Image.URL
		}
		rssFeed.Channel.Item[i] = RSSItem{
			Title:       item.Title,
			Link:        item.Link,
			Description: item.Description,
			PubDate:     item.Published,
			ImageURL:    imageURL,
		}
	}
}
