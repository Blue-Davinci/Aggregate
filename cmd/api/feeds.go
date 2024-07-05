package main

import (
	"errors"
	"net/http"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/google/uuid"
)

func (app *application) createFeedHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body to get the feed data
	var input struct {
		Name            string `json:"name"`
		Url             string `json:"url"`
		ImgURL          string `json:"img_url"`
		FeedType        string `json:"feed_type"`
		FeedDescription string `json:"feed_description"`
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

func (app *application) getAllFeedsFollowedHandler(w http.ResponseWriter, r *http.Request) {
	// make a struct to hold what we would want from the queries
	//
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
	// get the sort values falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Add the supported sort values for this endpoint to the sort safelist.
	input.Filters.SortSafelist = []string{"id", "name", "url", "-id", "-name", "-url"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	feed_follows, metadata, err := app.models.Feeds.GetAllFeedsFollowedByUser(app.contextGetUser(r).ID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Return the feeds in the response body
	err = app.writeJSON(w, http.StatusOK, envelope{"feed_follows": feed_follows, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFeedFollowHandler(w http.ResponseWriter, r *http.Request) {
	//Read our data as parameters from the URL as it's a HTTP DELETE Request
	feedfollowID, err := app.readIDParam(r)
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
