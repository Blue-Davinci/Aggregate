package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
	"github.com/felixge/httpsnoop"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.PrintInfo("Authenticating request", map[string]string{
			"host":            r.Host,
			"Request Method:": r.Method,
			"Request URI:":    r.RequestURI,
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

func (app *application) limitations(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user from the request context.
		user := app.contextGetUser(r)
		// now we check if the user has an active subscription, if they do, for now
		// we will just let them through and execute the next handler in the chain.
		app.logger.PrintInfo("Checking user limitations", map[string]string{
			"User ID": strconv.FormatInt(user.ID, 10),
		})
		// TODO: Get users subscription that is currently active or cancelled but not expired.
		_, err := app.models.Payments.GetActiveOrNonExpiredSubscriptionByID(user.ID)
		if err != nil {
			// if the user does not have a subscription...
			switch {
			case errors.Is(err, data.ErrSubscriptionNotFound):
				// this means they do not have a subscription and we should check if they have exceeded their limits.
				limitations, err := app.models.Limitations.GetUserLimitations(user.ID)
				if err != nil {
					app.serverErrorResponse(w, r, err)
					return
				}
				// check if the user has exceeded their feed limit
				if limitations.Created_Feeds >= int64(app.config.limitations.maxFeedsCreated) {
					app.limitationResponse(w, r)
					return
				}
				// check if the user has exceeded their followed feed limit
				if limitations.Followed_Feeds >= int64(app.config.limitations.maxFeedsFollowed) {
					app.limitationResponse(w, r)
					return
				}
				// check if the user has exceeded their comments limit
				if limitations.Comments_Today >= int64(app.config.limitations.maxComments) {
					app.limitationResponse(w, r)
					return
				}
			default:
				// this means there was an error and we should return a server error
				app.serverErrorResponse(w, r, err)
				return
			}
		}
		// if the user has a subscription or has not exceeded their limitations, we call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// The rateLimit() middleware will be used to rate limit the number of requests that a
// client can make to certain routes within a given time window.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// Define a client struct to hold the rate limiter and last seen time for each
	// client.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold the clients' IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)
	// Launch a background goroutine which removes old entries from the clients map once
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)
			// Lock the mutex to prevent any rate limiter checks from happening while
			// the cleanup is taking place.
			mu.Lock()
			// Loop through all clients. If they haven't been seen within the last three
			// minutes, delete the corresponding entry from the map.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Importantly, unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only carry out the check if rate limiting is enabled.
		if app.config.limiter.enabled {
			// Extract the client's IP address from the request.
			ip := realip.FromRequest(r)
			// Lock the mutex to prevent this code from being executed concurrently.
			mu.Lock()
			// Check to see if the IP address already exists in the map. If it doesn't, then
			// initialize a new rate limiter and add the IP address and limiter to the map.
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					// Use the requests-per-second and burst values from the config struct.
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}
			// Update the last seen time for the client.
			clients[ip].lastSeen = time.Now()

			// Call the Allow() method on the rate limiter for the current IP address. If
			// the request isn't allowed, unlock the mutex and send a 429 Too Many Requests
			// response, just like before.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			// unlock the mutex before calling the next handler in the
			// chain
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}
