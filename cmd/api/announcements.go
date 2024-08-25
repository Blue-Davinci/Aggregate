package main

import (
	"errors"
	"net/http"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
)

// GetAnnouncmentsForUser() method gets all the announcements for a user. It takes in a
// user ID and returns a slice of announcements and an error if there was an issue with the
// database query.
func (app *application) GetAnnouncmentsForUser(w http.ResponseWriter, r *http.Request) {
	// simple implementation and we just get all the announcement by ID
	announcements, err := app.models.Announcements.GetAnnouncmentsForUser(app.contextGetUser(r).ID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrAnnouncementNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// write the response
	err = app.writeJSON(w, http.StatusOK, envelope{"announcements": announcements}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// MarkAnnouncmentAsReadByUser() method marks an announcement as read by a user. It takes in
// input of the user_id and announcement_id and returns an announcement read struct and an error
func (app *application) MarkAnnouncmentAsReadByUser(w http.ResponseWriter, r *http.Request) {
	// input struct
	var input struct {
		AnnouncementID int32 `json:"announcement_id"`
	}
	// read the input
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// create the announcement read struct
	announcementRead := &data.AnnouncementRead{
		UserID:         app.contextGetUser(r).ID,
		AnnouncementID: input.AnnouncementID,
	}
	// validate
	v := validator.New()
	if data.ValidateReadAnnouncement(v, announcementRead); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// mark the announcement as read
	err = app.models.Announcements.MarkAnnouncmentAsReadByUser(announcementRead)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// write the response
	err = app.writeJSON(w, http.StatusCreated, envelope{"announcement_read": announcementRead}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
