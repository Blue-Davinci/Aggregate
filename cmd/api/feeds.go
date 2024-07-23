package main

import (
	"errors"
	"net/http"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

// createFeedHandler() creates a new feed record in the database. It expects the request body to contain
// a JSON object with the feed data. The handler reads the JSON data from the request body, validates it,
// and then uses the Insert() method on the feedModel to insert the new feed record into the database.
func (app *application) createFeedHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the feed data
	var input struct {
		Name            string `json:"name"`
		Url             string `json:"url"`
		ImgURL          string `json:"img_url"`
		FeedType        string `json:"feed_type"`
		FeedDescription string `json:"feed_description"`
		Is_Hidden       bool   `json:"is_hidden"`
	}
	//Read our data into the input struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	app.logger.PrintInfo("Creating a new feed...", map[string]string{
		"input": input.Name,
	})

	// Create a new Feed record in the database while using contextGetUser
	// To pass down the user information to the feed model
	feed := &data.Feed{
		Name:            input.Name,
		Url:             input.Url,
		ImgURL:          input.ImgURL,
		FeedType:        input.FeedType,
		FeedDescription: input.FeedDescription,
		UserID:          app.contextGetUser(r).ID,
		Is_Hidden:       input.Is_Hidden,
	}
	// Initialize a new Validator.
	v := validator.New()
	if data.ValidateFeed(v, feed); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Call the Insert() method on the feedModel to insert the feed record into the database.
	err = app.models.Feeds.Insert(feed)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateFeed):
			v.AddError("url", "This feed already exists")
			app.failedConstraintValidation(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 201 Created status code and the new feed record in the response body
	err = app.writeJSON(w, http.StatusCreated, envelope{"feed": feed}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getAllFeedsHandler() returns all the feeds that exist in our Database.
// This endpoint also facilitates or supports pagination data. We have support
// for name and URL queries as well as sorting, but still to impliment it.
func (app *application) getAllFeedsHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	var input struct {
		Name string
		Url  string
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	// use our helpers to convert the queries
	input.Name = app.readString(qs, "name", "")
	input.Url = app.readString(qs, "url", "")
	//get the page & pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "name")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"name", "url", "-name", "-url"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	//fmt.Println(">> Page and Pagesize: ", input.Filters.Page, input.Filters.PageSize)
	feeds, metadata, err := app.models.Feeds.GetAllFeeds(input.Name, input.Url, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"feeds": feeds, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GetListOfFollowedFeedsHandler() GET /feeds/followed, Returns the feeds followed by the user
func (app *application) getListOfFollowedFeedsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the page & pagesizes as ints and set to the embedded struct
	input.Name = app.readString(qs, "name", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// get the sort values falling back to "created_at" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "created_at")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"created_at", "-created_at"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	feeds, metadata, err := app.models.Feeds.GetListOfFollowedFeeds(app.contextGetUser(r).ID, input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"feeds": feeds, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// getTopFollowedFeedsHandler() gets the top 'x' followed feeds by grabbing the feed follows,
// sorting, filtering and grabing the top 'x' feeds returning it as a TopFeed struct.
// the default is 5 feeds, but a user can specify using the page_size query
func (app *application) getTopFollowedFeedsHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	//
	var input struct {
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the pagesizes as ints and set to the embedded struct
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 5, v)
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "-id"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	topFollowedFeeds, err := app.models.Feeds.GetTopFollowedFeeds(input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	app.logger.PrintInfo("Returning Top Followed Feeds", nil)
	err = app.writeJSON(w, http.StatusOK, envelope{"top_followed_feeds": topFollowedFeeds}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getAllFeedsFollowedHandler() GET /feeds/favorites, Returns the favorited feeds for the user
func (app *application) getAllFeedsFollowedHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	//we also need a member for name for search queries
	var input struct {
		Name string
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the page & pagesizes as ints and set to the embedded struct
	input.Name = app.readString(qs, "name", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "name")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"name", "-name"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	feed_follows, metadata, err := app.models.Feeds.GetAllFeedsFollowedByUser(app.contextGetUser(r).ID, input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"feeds": feed_follows, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// deleteFeedFollowHandler deletes a feed follow record from the database.
func (app *application) deleteFeedFollowHandler(w http.ResponseWriter, r *http.Request) {
	//Read our data as parameters from the URL as it's a HTTP DELETE Request
	feedfollowID, err := app.readIDParam(r, "feedID")
	//check whether there's an error or the feedID is invalid
	if err != nil || feedfollowID == uuid.Nil {
		app.notFoundResponse(w, r)
		return
	}
	// Create a new Feed record in the database while using contextGetUser
	// To pass down the user information to the feed model
	feedfollow := &data.FeedFollow{
		ID:     feedfollowID,
		UserID: app.contextGetUser(r).ID,
	}
	// Initialize a new Validator.
	v := validator.New()
	if data.ValidateFeedFollow(v, feedfollow); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Call delete to perform the unfollow operation
	err = app.models.Feeds.DeleteFeedFollow(feedfollow)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrFeedFollowNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "feed successfully unfollowed"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// createFeedFollowHandler creates a new feed follow record in the database. It expects the request body to contain
// a JSON object with the feed data. The handler reads the JSON data from the request body, validates it,
// and then uses the Insert() method on the feedModel to insert the new feed record into the database.
func (app *application) createFeedFollowHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the feed data
	var input struct {
		FeedId uuid.UUID `json:"feed_id"`
	}
	//Read our data into the input struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create a new Feed record in the database while using contextGetUser
	// To pass down the user information to the feed model
	feedfollow := &data.FeedFollow{
		ID:     input.FeedId,
		UserID: app.contextGetUser(r).ID,
	}
	// Initialize a new Validator.
	v := validator.New()
	if data.ValidateFeedFollow(v, feedfollow); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Call the Insert() method on the feedModel to insert the feed record into the database.
	follow, err := app.models.Feeds.CreateFeedFollow(feedfollow)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateFollow):
			v.AddError("follow", "cannot follow the same feed twice")
			app.failedConstraintValidation(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 201 Created status code and the new feed record in the response body
	err = app.writeJSON(w, http.StatusCreated, envelope{"feed_follow": follow}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GetFeedsCreatedByUserHandler() GET /feeds/created, Returns the feeds created by the user
func (app *application) getFeedsCreatedByUserHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	//we also need a member for name for search queries
	var input struct {
		Name string
		data.Filters
	}
	//validate if queries are provided
	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	//get the page & pagesizes as ints and set to the embedded struct
	input.Name = app.readString(qs, "name", "")
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "name")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"name", "-name"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	feeds, metadata, err := app.models.Feeds.GetAllFeedsCreatedByUser(app.contextGetUser(r).ID, input.Name, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"feeds": feeds, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateFeedHandler(w http.ResponseWriter, r *http.Request) {
	//Read our data as parameters from the URL as it's a HTTP PATCH Request
	feedID, err := app.readIDParam(r, "feedID")
	//check whether there's an error or the feedID is invalid
	if err != nil || feedID == uuid.Nil {
		app.notFoundResponse(w, r)
		return
	}
	// Get our feed.
	feed, err := app.models.Feeds.GetFeedByID(feedID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Have an input field that holds the feed data
	// We use a pointer to the string to differentiate between an empty string and a null value
	var input struct {
		Name            *string `json:"name"`
		Url             *string `json:"url"`
		ImgURL          *string `json:"img_url"`
		FeedType        *string `json:"feed_type"`
		FeedDescription *string `json:"feed_description"`
		Is_Hidden       bool    `json:"is_hidden"`
	}
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Check if the input fields are empty and if they are, we set them to the current values
	if input.Name != nil {
		feed.Name = *input.Name
	}
	// check for the url
	if input.Url != nil {
		feed.Url = *input.Url
	}
	// check for the img_url
	if input.ImgURL != nil {
		feed.ImgURL = *input.ImgURL
	}
	// check for the feed_type
	if input.FeedType != nil {
		feed.FeedType = *input.FeedType
	}
	// check for the feed_description
	if input.FeedDescription != nil {
		feed.FeedDescription = *input.FeedDescription
	}
	feed.Is_Hidden = input.Is_Hidden

	// Initialize a new Validator.
	v := validator.New()
	if data.ValidateFeed(v, feed); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Call the Update() method on the feedModel to update the feed record in the database.
	err = app.models.Feeds.UpdateFeed(feed)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code and the updated feed record in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"feed": feed}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
