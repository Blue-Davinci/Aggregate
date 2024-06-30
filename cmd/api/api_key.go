package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
)

func (app *application) createAuthenticationApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	//read the data from the request
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// validate the user's password & email
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// get the user from the database
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		// if the user is not found, we return an invalid credentials response
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			// otherwsie return a 500 internal server error
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// check if the password matches
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// if password doesn't match then we shout
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}
	// Otherwise, if the password is correct, we generate a new api_key with a 72-hour
	// expiry time and the scope 'authentication', saving it to the DB
	api_key, err := app.models.ApiKey.New(user.ID, 24*time.Hour, data.ScopeAuthentication, data.APIKeyLength)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Encode the apikey to json and send it to the user with a 201 Created status code
	err = app.writeJSON(w, http.StatusCreated, envelope{"api_key": api_key}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
