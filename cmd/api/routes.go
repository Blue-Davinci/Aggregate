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
		AllowedOrigins:   []string{"https://adapted-healthy-monitor.ngrok-free.app", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
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
	// The top routes will also need to be seperated when we add more, currently
	// top feeds is there and will be available to anyone.
	v1Router.Mount("/top", app.statisticRoutes())

	v1Router.Mount("/users", app.userRoutes(&dynamicMiddleware))
	v1Router.Mount("/feeds", app.feedRoutes(&dynamicMiddleware))
	v1Router.Mount("/search-options", app.searchOptionsRoutes(&dynamicMiddleware))
	v1Router.Mount("/api", app.apiKeyRoutes())
	v1Router.Mount("/subscriptions", app.subscriptionRoutes(&dynamicMiddleware))

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
func (app *application) userRoutes(dynamicMiddleware *alice.Chain) chi.Router {
	userRoutes := chi.NewRouter()
	userRoutes.Post("/", app.registerUserHandler)
	// /activation : for activating accounts
	userRoutes.Put("/activated", app.activateUserHandler)
	// **/password : for updating passwords.
	userRoutes.Put("/password", app.updateUserPasswordHandler)
	// update user info. This will be a dynamically protected route.
	userRoutes.With(dynamicMiddleware.Then).Patch("/", app.updateUserInformationHandler)
	return userRoutes
}

// feedRoutes() provides a router for the /feeds API endpoint.
// We pass the pointer to the dynamic middleware here because some
// Of the routes require verified and activated users
func (app *application) feedRoutes(dynamicMiddleware *alice.Chain) chi.Router {
	feedRoutes := chi.NewRouter()
	//authenticated/activated endpoints
	feedRoutes.With(dynamicMiddleware.Then).Post("/", app.createFeedHandler)
	// routes to get favorited posts, favorite and unfavorite posts as well.
	feedRoutes.With(dynamicMiddleware.Then).Get("/favorites", app.GetRSSFavoritePostsForUserHandler)
	feedRoutes.With(dynamicMiddleware.Then).Post("/favorites", app.CreateRSSFavoritePostHandler)
	feedRoutes.With(dynamicMiddleware.Then).Delete("/favorites/{postID}", app.DeleteFavoritePostHandler)

	feedRoutes.With(dynamicMiddleware.Then).Get("/favorites/posts", app.GetDetailedFavoriteRSSPosts)

	feedRoutes.With(dynamicMiddleware.Then).Get("/follow", app.getAllFeedsFollowedHandler)
	feedRoutes.With(dynamicMiddleware.Then).Get("/follow/list", app.getListOfFollowedFeedsHandler)

	feedRoutes.With(dynamicMiddleware.Then).Post("/follow", app.createFeedFollowHandler)
	feedRoutes.With(dynamicMiddleware.Then).Delete("/follow/{feedID}", app.deleteFeedFollowHandler)
	feedRoutes.With(dynamicMiddleware.Then).Get("/follow/posts", app.getFollowedRssPostsForUserHandler)
	feedRoutes.With(dynamicMiddleware.Then).Get("/follow/posts/{postID}", app.getRSSFeedByIDHandler)
	feedRoutes.With(dynamicMiddleware.Then).Post("/follow/posts/comments", app.createCommentHandler)

	feedRoutes.With(dynamicMiddleware.Then).Get("/follow/posts/comments/{postID}", app.getCommentsForPostHandler)
	feedRoutes.With(dynamicMiddleware.Then).Patch("/follow/posts/comments", app.updateUserCommentHandler)
	feedRoutes.With(dynamicMiddleware.Then).Delete("/follow/posts/comments/{commentID}", app.deleteCommentHandler)

	feedRoutes.With(dynamicMiddleware.Then).Delete("/follow/posts/comments/notifications/{postID}", app.deleteReadCommentNotificationHandler)

	feedRoutes.With(dynamicMiddleware.Then).Get("/created", app.getFeedsCreatedByUserHandler)
	feedRoutes.With(dynamicMiddleware.Then).Patch("/created/{feedID}", app.updateFeedHandler)

	//A general route that will serve as one of the public endpoints/"Home"
	feedRoutes.Get("/", app.getAllFeedsHandler)
	feedRoutes.Get("/{feedID}", app.getFeedWithStatsHandler)
	feedRoutes.Get("/sample-posts/{feedID}", app.getRandomRSSPostsHandler)

	return feedRoutes
}

func (app *application) searchOptionsRoutes(dynamicMiddleware *alice.Chain) chi.Router {
	searchOptionsRoutes := chi.NewRouter()
	searchOptionsRoutes.With(dynamicMiddleware.Then).Get("/feeds", app.getFeedSearchOptionsHandler)
	// This is a general route intended for the feeds search options
	searchOptionsRoutes.Get("/feed-types", app.getFeedTypeSearchOptionsHandler)
	return searchOptionsRoutes
}

func (app *application) statisticRoutes() chi.Router {
	metricRoutes := chi.NewRouter()
	metricRoutes.Get("/feeds", app.getTopFollowedFeedsHandler)
	metricRoutes.Get("/creators", app.getTopFeedCreatorsHandler)
	return metricRoutes
}

// apiKeyRoutes() provides a router for the /api API endpoint.
// That is, it is responsible for the generation of the API Keys to
// users who should be: Signup and Activated
func (app *application) apiKeyRoutes() chi.Router {
	apiKeyRoutes := chi.NewRouter()
	apiKeyRoutes.Post("/authentication", app.createAuthenticationApiKeyHandler)
	// initial request for token
	apiKeyRoutes.Post("/password-reset", app.createPasswordResetTokenHandler)
	// /password-reset : for sending keys for resetting passwords

	return apiKeyRoutes
}

// subscriptionRoutes() provides a router for the /subscriptions API endpoint.
// It is responsible for the subscription/paments for users
func (app *application) subscriptionRoutes(dynamicMiddleware *alice.Chain) chi.Router {
	subscriptionRoutes := chi.NewRouter()
	subscriptionRoutes.With(dynamicMiddleware.Then).Post("/initialize", app.initializeTransactionHandler)
	subscriptionRoutes.With(dynamicMiddleware.Then).Post("/verify", app.verifyTransactionHandler)
	// plans is free to everyone
	subscriptionRoutes.Get("/plans", app.getPaymentPlansHandler)
	return subscriptionRoutes
}
