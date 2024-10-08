package main

import "net/http"

// getFeedSearchOptionsHandler() Is a search option endpoint designed to return all available
// distinc feeds from the DB
func (app *application) getFeedSearchOptionsHandler(w http.ResponseWriter, r *http.Request) {

	// Get the feed search options from the database
	searchOptions, err := app.models.SearchOptions.GetFeedSearchOptions()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//return the search options as a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"feeds": searchOptions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getFeedTypeSearchOptionsHandler() Is a search option endpoint designed to return all available
// distinc feed types from the DB
func (app *application) getFeedTypeSearchOptionsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the feed type search options from the database
	searchOptions, err := app.models.SearchOptions.GetFeedTypeSearchOptions()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//return the search options as a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"feed_types": searchOptions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getFeedPrioritySearchOptionsHandler() Is a search option endpoint designed to return all available
// distinc feed priorities from the DB
func (app *application) getFeedPrioritySearchOptionsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the feed priority search options from the database
	searchOptions, err := app.models.SearchOptions.GetFeedPrioritySearchOptions()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//return the search options as a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"feed_priorities": searchOptions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getGetErrorTypeSearchOptionsHandler() Is a search option endpoint designed to return all available
// distinc error types from the DB
func (app *application) getGetErrorTypeSearchOptionsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the feed priority search options from the database
	searchOptions, err := app.models.SearchOptions.GetErrorTypeSearchOptions()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//return the search options as a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"error_types": searchOptions}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
