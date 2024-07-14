package main

import "net/http"

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
