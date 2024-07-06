package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/justinas/alice"
)

// routes() provides a router for the API. It is responsible for
// mounting all the routes we have in the application.
func (app *application) routes() http.Handler {
	router := chi.NewRouter()
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	//Use alice to make a global middleware chain.
	globalMiddleware := alice.New(app.recoverPanic, app.metrics, app.authenticate).Then
	// Dynamic Middleware, these will apply to only select routes
	dynamicMiddleware := alice.New(app.requireAuthenticatedUser, app.requireActivatedUser)
	router.Use(globalMiddleware)
	// Make our categorized routes
	v1Router := chi.NewRouter()
	// Mounts general routes "home"
	v1Router.With(dynamicMiddleware.Then).Mount("/", app.generalRoutes())
	v1Router.Mount("/users", app.userRoutes())
	v1Router.Mount("/feeds", app.feedRoutes(&dynamicMiddleware))
	v1Router.Mount("/api", app.apiKeyRoutes())

	// Mount to our base version
	router.Mount("/v1", v1Router)
	return router
}

// generalRoutes() provides a router for the general routes.
// Mounted rirectly after our version url. They contaon sanity and
// health checks. Probably add other AOB's here.
func (app *application) generalRoutes() chi.Router {
	generalRoutes := chi.NewRouter()
	// /debug/vars : for expvar, wrapping it in a handler func for assertion otherwise it complains
	generalRoutes.Get("/debug/vars", func(w http.ResponseWriter, r *http.Request) {
		expvar.Handler().ServeHTTP(w, r)
	})
	generalRoutes.Get("/health", app.healthcheckHandler)
	generalRoutes.Get("/notifications", app.getUserNotificationsHandler)
	return generalRoutes
}

// userRoutes() provides a router for the /users API endpoint.
// We pass the pointer to the dynamic middleware here because some
// Of the routes require verified and activated users while some don't
func (app *application) userRoutes() chi.Router {
	userRoutes := chi.NewRouter()
	userRoutes.Post("/", app.registerUserHandler)
	// /activation : for activating accounts
	userRoutes.Put("/activated", app.activateUserHandler)
	// /password-reset : for resetting passwords
	return userRoutes
}

// feedRoutes() provides a router for the /feeds API endpoint.
// We pass the pointer to the dynamic middleware here because some
// Of the routes require verified and activated users
func (app *application) feedRoutes(dynamicMiddleware *alice.Chain) chi.Router {
	feedRoutes := chi.NewRouter()
	//authenticated/activated endpoints
	feedRoutes.With(dynamicMiddleware.Then).Post("/", app.createFeedHandler)
	feedRoutes.With(dynamicMiddleware.Then).Post("/follow", app.createFeedFollowHandler)
	feedRoutes.With(dynamicMiddleware.Then).Get("/follow", app.getAllFeedsFollowedHandler)
	feedRoutes.With(dynamicMiddleware.Then).Delete("/follow/{feedID}", app.deleteFeedFollowHandler)
	feedRoutes.With(dynamicMiddleware.Then).Get("/follow/posts", app.GetFollowedRssPostsForUserHandler)

	//A general route that will serve as one of the public endpoints/"Home"
	feedRoutes.Get("/", app.getAllFeedsHandler)

	return feedRoutes
}

// apiKeyRoutes() provides a router for the /api API endpoint.
// That is, it is responsible for the generation of the API Keys to
// users who should be: Signup and Activated
func (app *application) apiKeyRoutes() chi.Router {
	apiKeyRoutes := chi.NewRouter()
	apiKeyRoutes.Post("/authentication", app.createAuthenticationApiKeyHandler)
	// /password-reset : for sending keys for resetting passwords

	return apiKeyRoutes
}
