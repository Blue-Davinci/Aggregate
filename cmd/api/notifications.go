package main

import (
	"fmt"
	"net/http"
)

func (app *application) fetchNotificationsHandler() {
	app.logger.PrintInfo("Starting notification fetch job...", nil)
	_, err := app.config.notifier.cronJob.AddFunc("*/1 * * * *", app.startNotificationFetch)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"Error": "Error adding running notifier job",
		})
	}
	// start the cron scheduler
	app.config.notifier.cronJob.Start()
}

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

func (app *application) getUserNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the request context
	userID := app.contextGetUser(r).ID
	// Get the interval from the request context
	interval := app.config.notifier.interval
	// Get the notifications for the user
	notifications, err := app.models.Notifications.GetUserNotifications(userID, interval)
	if err != nil {
		app.logger.PrintError(err, nil)
	}
	// Send the notifications to the client
	err = app.writeJSON(w, http.StatusOK, envelope{"notifications": notifications}, nil)
	if err != nil {
		app.logger.PrintError(err, nil)
	}
}
