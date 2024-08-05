package main

import "net/http"

func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.RequestURI,
		"ip_address":     r.RemoteAddr,
	})
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env, nil)

	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)
	}
}
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// The badRequestResponse() method will be used to send a 400 Bad Request status code and
// JSON response to the client.
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// Note that the errors parameter here has the type map[string]string, which is exactly
// the same as the errors map contained in our Validator type.
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// The invalidCredentialsResponse() method will return invalid token credential error
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// The failedConstraintValidation() method will be used to send a 409 Conflict status code and
// JSON response to the client incase of DB general constraint violations.
func (app *application) failedConstraintValidation(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusConflict, errors)
}

// The editConflictResponse() method will be used to send a 409 Conflict status code and
// JSON response to the client incase of errors that arise during updating of records.
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}

// The invalidAuthenticationTokenResponse() method will return invalid token error
// we set the header to show the client that the apikey is required/Invalid
func (app *application) invalidAuthenticationApiResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "ApiKey")
	message := "invalid or missing authentication api key"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// The authenticationRequiredResponse() method will be used to send a 401 Unauthorized status if user
// is not authenticated
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

// The inactiveAccountResponse() method will be used to send a 403 Forbidden status if user
// is authenticated but not activated
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

// The limitationResponse() method will be used to send a 403 Forbidden status if user
// is authenticated and not subscribed but has reached their limit for a particular action
func (app *application) limitationResponse(w http.ResponseWriter, r *http.Request) {
	message := "you have reached your unsubscribed limit for this action"
	app.errorResponse(w, r, http.StatusForbidden, message)
}

// The notFoundResponse() method will be used to send a 404 Not Found status code and
// JSON response to the client.
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}
