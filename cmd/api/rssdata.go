package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

// startRssFeedScraperHandler() Is the entry point of our scraper function
// It Uses noofroutines and fetchinterval settings from our config then
// Proceeds to get the feeds to fetch, summoning the Main scraper.
func (app *application) startRssFeedScraperHandler() {
	goroutines := app.config.scraper.noofroutines
	interval := app.config.scraper.fetchinterval
	app.logger.PrintInfo("Starting RSS Feed Scraper", map[string]string{
		"No of Go Routines": fmt.Sprintf("%d", goroutines),
		"Interval":          fmt.Sprintf("%ds", interval),
		"No of Retries":     fmt.Sprintf("%d", app.config.scraper.scraperclient.retrymax),
		"Client Timeout":    fmt.Sprintf("%d", app.config.scraper.scraperclient.timeout),
	})
	// start the scraper
	// convert the interval to seconds
	ticker := time.NewTicker(time.Duration(interval) * time.Second)

	for ; ; <-ticker.C {
		feeds, err := app.models.RSSFeedData.GetNextFeedsToFetch(goroutines, interval)
		// if we get an error, we log it and continue wuth our work
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"Error Getting Feeds From DB": "GetNextFeedsToFetch",
			})
			continue
		}

		// For each particular feed, we pass the data to our main Scraping
		// function which launches  seperate go routines for the work.
		app.logger.PrintInfo("Starting scraping workers", map[string]string{
			"Executing workers": fmt.Sprintf("Getting %d feeds", len(feeds)),
		})
		for _, feed := range feeds {
			app.rssFeedScraper(feed)
		}
	}

}

// rssFeedScraper() is the main method which performs scraping for each
// individual feed. It takes in an indvidiual Feed, updates its last fetched
// using MarkFeedAsFetched() and then saved the data to our DB
func (app *application) rssFeedScraper(feed database.Feed) {
	// we want to fetch each of the feeds concurrently, so we make a wait group
	// using our app.background(func(){}) through a for loop to iterate over the feeds starting a routine for each feed
	app.background(func() {
		// get the feed data
		err := app.models.RSSFeedData.MarkFeedAsFetched(feed.ID)
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"Error Marking Feed As Fetched": "MarkFeedAsFetched",
				"Feed Name":                     feed.Name,
				"Feed ID":                       feed.ID.String(),
			})
			return
		}
		// call our GetRSSFeeds to return all feeds for each specific URL
		rssFeeds, err := app.models.RSSFeedData.GetRSSFeeds(
			app.config.scraper.scraperclient.retrymax,
			app.config.scraper.scraperclient.timeout,
			feed.Url)
		if err != nil {
			switch {
			case err == data.ErrContextDeadline:
				app.logger.PrintInfo(err.Error(), map[string]string{
					"Feed": feed.Name,
					"URL":  feed.Url,
				})
			case err == data.ErrUnableToDetectFeedType:
				app.logger.PrintInfo(err.Error(), map[string]string{
					"Feed": feed.Name,
					"URL":  feed.Url,
				})
			default:
				app.logger.PrintError(err, map[string]string{
					"Error Fetching RSS Feeds": "GetRSSFeeds",
					"URL":                      feed.Url,
				})
			}
			if err == data.ErrContextDeadline {

				return
			}
		}
		// store the fetched data into our DB
		err = app.models.RSSFeedData.CreateRssFeedPost(&rssFeeds, &feed.ID)
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"Error Creating Rss Feed Post": "CreateRssFeedPost",
				"Feed Name":                    feed.Name})
			return
		}

		/*app.logger.PrintInfo("Finished collecting feeds for: ", map[string]string{
			"Name":   feed.Name,
			"Posts:": fmt.Sprintf("%d", len(rssFeeds.Channel.Item)),
		})*/
	})
}

// Handler for out GetAllPost
func (app *application) GetFollowedRssPostsForUserHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.PrintInfo("Getting Followed RSS Posts for User", nil)
	// make a struct to hold what we would want from the queries
	var input struct {
		Name    string
		Feed_ID uuid.UUID
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	// get our parameters
	input.Name = app.readString(qs, "name", "")      // get our name parameter
	feed_id, err := app.readIDFromQuery(r, "feedID") // get our feed_id parameter
	// if no FEED ID is provided, we expressly set it to nil so that our
	// query identifies it as a nil value. Our app never gives a user nil
	// uuid's so we can be sure that if we get a nil value, it is because
	// the user did not provide a feed_id
	if err != nil || feed_id == uuid.Nil {
		input.Feed_ID = uuid.Nil
	} else {
		input.Feed_ID = feed_id
	}
	//get the page & pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// We don't use any sort for this endpoint
	input.Filters.Sort = app.readString(qs, "", "")
	// None of the sort values are supported for this endpoint
	input.Filters.SortSafelist = []string{"", ""}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// We are good, now we call our models getposts to get the rss posts
	userRssFollowedPosts, metadata, err := app.models.RSSFeedData.GetFollowedRssPostsForUser(app.contextGetUser(r).ID,
		input.Name,
		input.Feed_ID,
		input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"followed_rss_posts": userRssFollowedPosts, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) getRandomRSSPostsHandler(w http.ResponseWriter, r *http.Request) {
	//Read our data as parameters from the URL as it's a HTTP DELETE Request
	feedID, err := app.readIDParam(r, "feedID")
	//check whether there's an error or the feedID is invalid
	if err != nil || feedID == uuid.Nil {
		app.notFoundResponse(w, r)
		return
	}
	// we will use the feedfollow verification as we are verifying the same thing
	feedfollow := &data.FeedFollow{
		ID: feedID,
	}
	v := validator.New()
	if data.ValidateFeedFollow(v, feedfollow); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get the pagination data
	var input struct {
		data.Filters
	}
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 3, v)
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "-id"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// now get our posts
	rssPosts, err := app.models.RSSFeedData.GetRandomRSSPosts(feedID, input.Filters)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPostNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"rss_posts": rssPosts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// getRSSFeedByIDHandler returns a single RSS Feed Post with favorite by its ID
func (app *application) getRSSFeedByIDHandler(w http.ResponseWriter, r *http.Request) {
	postIDValue, err := app.readIDParam(r, "postID")
	if err != nil || postIDValue == uuid.Nil {
		app.notFoundResponse(w, r)
		return
	}
	// Create a new favorite post to read in the data
	favoritePost := &data.RSSPostFavorite{
		Post_ID: postIDValue,
	}
	// Initialize a new Validator.
	v := validator.New()
	if data.ValidatePostID(v, favoritePost); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	rssFeedPost, err := app.models.RSSFeedData.GetRSSFeedByID(app.contextGetUser(r).ID, postIDValue)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrPostNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"rss_post": rssFeedPost}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// GetRSSFavoritePostsForUserHandler Gets all posts that have been marked as favorites by a user
// This is a GET request to /feeds/favorites and passes through the user-id
func (app *application) GetRSSFavoritePostsForUserHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	var input struct {
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the page & pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// We don't use any sort for this endpoint
	input.Filters.Sort = app.readString(qs, "", "")
	// None of the sort values are supported for this endpoint
	input.Filters.SortSafelist = []string{"", ""}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// We are good, now we call our models getposts to get the rss posts
	userRssFavoritePosts, err := app.models.RSSFeedData.GetRSSFavoritePostsForUser(app.contextGetUser(r).ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"favorite_rss_posts": userRssFavoritePosts}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// CreateRSSFavoritePostHandler Creates a new favorite post for a user
func (app *application) CreateRSSFavoritePostHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Post_ID uuid.UUID `json:"post_id"`
		Feed_ID uuid.UUID `json:"feed_id"`
	}
	// Read the JSON data from the request body
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	app.logger.PrintInfo("Creating a new favorite Post...", map[string]string{
		"input": input.Post_ID.String(),
	})
	// Get our context user
	user := app.contextGetUser(r)
	// Create a new feavorite post to read in the data
	favoritePost := &data.RSSPostFavorite{
		Post_ID: input.Post_ID,
		Feed_ID: input.Feed_ID,
	}
	// Initialize a validator for our data
	v := validator.New()
	if data.ValidateFavoritePost(v, favoritePost); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// check if the post exists
	err = app.models.RSSFeedData.CreateRSSFavoritePost(user.ID, favoritePost)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateFavorite):
			v.AddError("post_id", "cannot favorite a post twice")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 201 Created status code and the new Feed Favorited record in the response body
	err = app.writeJSON(w, http.StatusCreated, envelope{"favorite_rss_post": favoritePost}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteFeedFollowHandler Deletes a Post from the user's favorite list
// will accept a parameterized URL with the POST_ID
func (app *application) DeleteFavoritePostHandler(w http.ResponseWriter, r *http.Request) {
	// Read the postID from the URL
	postIDValue, err := app.readIDParam(r, "postID")
	if err != nil || postIDValue == uuid.Nil {
		app.notFoundResponse(w, r)
		return
	}
	// Create a new favorite post to read in the data
	favoritePost := &data.RSSPostFavorite{
		Post_ID: postIDValue,
	}
	// Initialize a new Validator.
	v := validator.New()
	if data.ValidatePostID(v, favoritePost); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Get our context user
	user := app.contextGetUser(r)
	// Call our delete function
	err = app.models.RSSFeedData.DeleteRSSFavoritePost(user.ID, favoritePost)
	if err != nil {
		app.logger.PrintInfo("2. In Deleter... ", nil)
		switch {
		case errors.Is(err, data.ErrPostNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "post successfully unfavorite"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GetRSSFavoritePostsOnlyForUserHandler() handles requests to get only favorite posts for a user
// It is a GET request to /feeds/favorites/posts
func (app *application) GetDetailedFavoriteRSSPosts(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string
		Feed_ID uuid.UUID
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	// get our parameters
	input.Name = app.readString(qs, "name", "")      // get our name parameter
	feed_id, err := app.readIDFromQuery(r, "feedID") // get our feed_id parameter
	// if no FEED ID is provided, we expressly set it to nil so that our
	// query identifies it as a nil value. Our app never gives a user nil
	// uuid's so we can be sure that if we get a nil value, it is because
	// the user did not provide a feed_id
	if err != nil || feed_id == uuid.Nil {
		input.Feed_ID = uuid.Nil
	} else {
		input.Feed_ID = feed_id
	}
	//get the page & pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// We don't use any sort for this endpoint
	input.Filters.Sort = app.readString(qs, "", "")
	// None of the sort values are supported for this endpoint
	input.Filters.SortSafelist = []string{"", ""}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// get the data

	favoritePosts, metadata, err := app.models.RSSFeedData.GetRSSFavoritePostsOnlyForUser(app.contextGetUser(r).ID, input.Name, input.Feed_ID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"favorite_rss_posts": favoritePosts, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
