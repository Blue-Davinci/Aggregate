package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/felixge/httpsnoop"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.PrintInfo("Authenticating request", map[string]string{
			"host": r.Host,
		})
		// add the vary header to indicate to any caches that the response varies
		// depending on the Authorization header.
		w.Header().Add("Vary", "Authorization")
		// Get the authorization header from the request. Will be empty if not found.
		authorizationHeader := r.Header.Get("Authorization")
		// If the Authorization header is empty, call contextStUser() to add an anonymous user.
		// Then call the next handler in the chain.
		if authorizationHeader == "" {
			fmt.Println("No Authorization Header")
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}
		// The header should look like this AUTHORIZATION: ApiKey <xxxxxxxxxx>, So we split it
		// into two parts using the space as the delimiter.
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "ApiKey" {
			app.invalidAuthenticationApiResponse(w, r)
			return
		}
		//extract the api key from the parts
		apikey := headerParts[1]
		// Validate the key
		v := validator.New()
		if data.ValidateAPIKeyPlaintext(v, apikey, data.APIVerificationLength); !v.Valid() {
			fmt.Println("Invalid API Key")
			app.invalidAuthenticationApiResponse(w, r)
			return
		}
		// Retrieve the details of the user associated with the authentication token,
		// again calling the invalidAuthenticationTokenResponse().
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, apikey)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				fmt.Println(">>Invalid API Key")
				app.invalidAuthenticationApiResponse(w, r)
			default:
				fmt.Println(">>>Invalid API Key")
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// Call the contextSetUser() helper to add the user information to the request
		// context.
		r = app.contextSetUser(r, user)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Create a new requireAuthenticatedUser() middleware to check that a user is not
// anonymous.
func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the contextGetUser() helper to retrieve the user
		// information from the request context.
		user := app.contextGetUser(r)
		// If the user is anonymous, then call the authenticationRequiredResponse() to
		// inform the client that they should authenticate before trying again.
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Checks that a user is both authenticated and activated.
func (app *application) requireActivatedUser(next http.Handler) http.Handler {
	// Rather than returning this http.HandlerFunc we assign it to the variable fn.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		// If the user is not activated, use the inactiveAccountResponse() helper to
		// inform them that they need to activate their account.
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Defer a function to recover any panic, and to log a message with a stack trace.
		defer func() {
			if err := recover(); err != nil {
				//set a header to triger Go's http Server to automatically close the connection
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {
	// Initialize the new expvar variables when the middleware chain is first built.
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_μs")
	// This will hold the response codes themselves
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use the Add() method to increment the number of requests received by 1.
		totalRequestsReceived.Add(1)
		// Use httpsnoop to capture the response statuses and executing the next handler in the chain
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		// On the way back up the middleware chain, increment the number of responses sent by 1.
		totalResponsesSent.Add(1)
		// Get response time from httpsnoop and increment the total processing time by that amount.
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())
		// Increment the number of responses sent with the status code.
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)

	})
}
