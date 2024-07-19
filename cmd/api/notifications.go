package main

import (
	"fmt"
	"net/http"

	"github.com/blue-davinci/aggregate/internal/data"
	"github.com/blue-davinci/aggregate/internal/validator"
)

// fetchNotificationsHandler() is a method that starts cron jobs for the:
// 1. notification fetch job - fetches all notifications based on the subset of feed names from fetched RSS posts
// 2. notification deleter job - clears old notifications
func (app *application) fetchNotificationsHandler() {
	app.logger.PrintInfo("Starting notification fetch job...", nil)
	// set our cron job interval to the interval set by the notifier-interval flag
	// or the app.config.notifier.interval struct field
	fetchIntervalString := fmt.Sprintf("*/%d * * * *", app.config.notifier.interval)
	_, err := app.config.notifier.cronJob.AddFunc(fetchIntervalString, app.startNotificationFetch)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"Error": "Error adding running notifier job",
		})
	}
	// start the notification deleter job to clear old notifications
	// default runs every 100 minutes, and can be changed via the notifier-delete-interval flag
	// or the app.config.notifier.deleteinterval struct field
	deleteIntervalString := fmt.Sprintf("*/%d * * * *", app.config.notifier.deleteinterval)
	_, err = app.config.notifier.cronJob.AddFunc(deleteIntervalString, app.clearOldNotificationsHandler)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"Error": "Error adding deleter notifier job",
		})
	}
	// start the cron scheduler
	app.config.notifier.cronJob.Start()
}

// startNotificationFetch() is a method that fetches all notifications based on the subset of feed names from fetched RSS posts
// we use the same interval as the notifier-interval flag or the app.config.notifier.interval struct field which are used
// to set the cron job interval i.e every cron job inherits the interval set to the configs
// this interval is used to check the latest posts. The default is 10min, so the job will run each and every 10 minutes
// fetching the latest 10 minutes posts and so forth. Changes to the config is reflected here as well.
func (app *application) startNotificationFetch() {
	// Fetch all feeds
	app.logger.PrintInfo("Running notification Worker...", nil)
	notifications, err := app.models.Notifications.FetchAndStoreNotifications(app.config.notifier.interval)
	if err != nil {
		app.logger.PrintError(err, nil)
	}
	app.logger.PrintInfo("Fetched new notifications", nil)
	for _, notification := range notifications {
		app.logger.PrintInfo("Notification", map[string]string{
			"Feed ID":   notification.Feed_ID.String(),
			"Feed Name": notification.Feed_Name,
		})
	}
	// After we are done fetching then, we loop through the bunches, saving
	// each notification to the database
	for _, notification := range notifications {
		notificationID, err := app.models.Notifications.InsertNotifications(notification)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
		app.logger.PrintInfo("Inserted notification", map[string]string{
			"Notification ID": fmt.Sprintf("%d", notificationID),
		})
	}
}

// clearOldNotificationsHandler() is a method that clears old notifications from the database
// Old notifications are those that are older than the delete interval set by the notifier-delete-interval flag
// or the app.config.notifier.deleteinterval struct field. The default is 100 minutes, so the job will run each and every 100 minutes
// deleting old notifications that are older than 100 minutes. Changes to the config is reflected here as well.
func (app *application) clearOldNotificationsHandler() {
	app.logger.PrintInfo("Running deleter Worker...", nil)
	// Delete old notifications
	err := app.models.Notifications.ClearOldNotifications(app.config.notifier.deleteinterval)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"Error deleting old notifications": "Error deleting old notifications",
		})
	}
	app.logger.PrintInfo("Deleted old notifications", nil)
}

func (app *application) deleteReadCommentNotificationHandler(w http.ResponseWriter, r *http.Request) {
	notificationID, err := app.readIDIntParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// we are okay now we proceed to delete the comment
	err = app.models.Notifications.DeleteReadCommentNotification(notificationID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Prepare a message
	message := fmt.Sprintf("comment notification with ID %d deleted", notificationID)
	// Return a 200 OK status code along with the deleted notification
	err = app.writeJSON(w, http.StatusOK, envelope{"message": message}, nil)
}

// getUserNotificationsHandler() is an endpoint function that retrieves all notifications
// we expect an interval to be passed as a parameter eg: /notifications?interval=10
// i[f none is passed, we default to the notifier-interval flag or the app.config.notifier.interval struct field]
func (app *application) getUserNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Interval int64
		data.Filters
	}

	v := validator.New()
	// Call r.URL.Query() to get the url.Values map containing the query string data.
	qs := r.URL.Query()
	// use our helpers to convert the queries
	input.Interval = int64(app.readInt(qs, "interval", int(app.config.notifier.interval), v))

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"interval", "id", "-id", "-interval"}
	// Perform validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Get the user ID from the request context
	userID := app.contextGetUser(r).ID
	// Get the interval from the request context
	interval := app.refineInterval(&input.Interval)
	// Get the notifications for the user
	notifications, err := app.models.Notifications.GetUserNotifications(userID, interval)
	if err != nil {
		app.logger.PrintError(err, nil)
	}
	app.logger.PrintInfo("Fetching user notifications", map[string]string{
		"notifications total":         fmt.Sprintf("%d", len(notifications.Notification)),
		"comment notifications total": fmt.Sprintf("%d", len(notifications.CommentNotification)),
	})
	// Send the notifications to the client
	err = app.writeJSON(w, http.StatusOK, envelope{"notification_group": notifications}, nil)
	if err != nil {
		app.logger.PrintError(err, nil)
	}
}

// refineInterval() is a method that refines the interval passed as a parameter
func (app *application) refineInterval(interval *int64) int64 {
	// Check if interval is nil or out of the valid range
	// There is no need to set an interval higher than the delete interval as
	// the notifications will have been deleted
	if interval == nil || *interval <= 0 || *interval > app.config.notifier.deleteinterval {
		return app.config.notifier.interval
	}
	return *interval
}
