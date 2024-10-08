package main

import (
	"errors"
	"fmt"
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
	api_key, err := app.models.ApiKey.New(user.ID, 72*time.Hour, data.ScopeAuthentication, data.APIKeyLength)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// we also get the roles for the user
	permissions, err := app.models.Permissions.GetAllPermissionsForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	userRole := "user"
	if permissions.Include(data.PermissionAdminRead) {
		userRole = "admin"
	}
	app.logger.PrintInfo(fmt.Sprintf("Permissions: %v", permissions), nil)
	// Encode the apikey to json and send it to the user with a 201 Created status code
	err = app.writeJSON(w, http.StatusCreated, envelope{
		"api_key": api_key,
		"user":    user,
		"role":    userRole,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// Generate a password reset token and send it to the user's email address.
func (app *application) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the user's email address.
	var input struct {
		Email string `json:"email"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Try to retrieve the corresponding user record for the email address. If it can't
	// be found, return an error message to the client.
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		// We willl use a generic error message to avoid leaking information about which
		// email addresses are registered with the system.
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("generic", "if we found a matching email address, we have sent password reset instructions to it")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return an error message if the user is not activated.
	if !user.Activated {
		v.AddError("email", "user account must be activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Otherwise, create a new password reset token with a 45-minute expiry time.
	token, err := app.models.ApiKey.New(user.ID, 45*time.Minute, data.ScopePasswordReset, data.TokenKeyLength)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Email the user with their password reset token.
	app.background(func() {
		data := map[string]any{
			"passwordResetURL":   app.config.frontend.passwordreseturl + token.Plaintext,
			"passwordResetToken": token.Plaintext,
		}
		// Since email addresses MAY be case sensitive, notice that we are sending this
		// email using the address stored in our database for the user --- not to the
		// input.Email address provided by the client in this request.
		err = app.mailer.Send(user.Email, "token_password_reset.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	// Send a 202 Accepted response and confirmation message to the client.
	// But use a generic message as well
	// an email will be sent to you containing password reset instructions
	env := envelope{"message": "if we found a matching email address, we have sent password reset instructions to it"}
	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// This is a "manual" function to resend an activation code incase the user didn't receive it.
func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse and validate the user's email address.
	var input struct {
		Email string `json:"email"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Try to retrieve the corresponding user record for the email address. If it can't
	// be found, return an error message to the client.
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("email", "no matching email address found")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return an error if the user has already been activated.
	if user.Activated {
		v.AddError("email", "user has already been activated")
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Otherwise, create a new activation token.
	token, err := app.models.ApiKey.New(user.ID, 3*24*time.Hour, data.ScopeActivation, data.TokenKeyLength)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Email the user with their additional activation token.
	app.background(func() {
		data := map[string]any{
			"reactivationURL": app.config.frontend.activationurl + token.Plaintext,
			"activationToken": token.Plaintext,
			"userName":        user.Name,
		}
		// Since email addresses MAY be case sensitive, notice that we are sending this
		// email using the address stored in our database for the user --- not to the
		// input.Email address provided by the client in this request.
		err = app.mailer.Send(user.Email, "token_activation.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})
	// Send a 202 Accepted response and confirmation message to the client.
	env := envelope{"message": "an email will be sent to you containing activation instructions"}
	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
