package data

import (
	"context"
	"fmt"
	"time"

	"github.com/blue-davinci/aggregate/internal/database"
	"github.com/google/uuid"
)

type NotificationsModel struct {
	DB *database.Queries
}

// Thi struct represents the entirety of what our notifications look like
// we will return a notification group currently made up of a Post Notification
// and a Comment Notification
type NotificationsGroup struct {
	Notification        []*Notification
	CommentNotification []*CommentNotification
}

type Notification struct {
	ID         int64     `json:"id"`
	Feed_ID    uuid.UUID `json:"feed_id"`
	Feed_Name  string    `json:"feed_name"`
	Post_Count int       `json:"post_count"`
	Created_At time.Time `json:"created_at"`
}

type CommentNotification struct {
	ID         int64     `json:"id"`
	Comment_ID uuid.UUID `json:"comment_id"`
	Post_ID    uuid.UUID `json:"post_id"`
	User_ID    int64     `json:"user_id"`
	Created_At time.Time `json:"created_at"`
}

// FetchAndStoreNotifications() is the notifier's main function which
// fetches all notifications based on the subset of feed names from fetched
// RSS posts. That is, the query will check latest posts, specified by interval,
// aggregate these posts by the  parent i.e feeds table's name and count the number
// of each feed's occurrence. The result is a slice of Notifications that look like
// this:
//
//	[{d54414ed-a09f-42c6-9d5c-cebcf04fb848 Engadget 50}
//	{ccabe4bd-97da-4454-900d-0e7f00bc59d6 BBC Cricket 36}
//	{62dfa525-3e6e-428c-8694-80ded8a71b0a Megaphone Podcast 3}]
func (m *NotificationsModel) FetchAndStoreNotifications(interval int64) ([]*Notification, error) {
	// Create a new context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Fetch all feeds
	notificationgroup, err := m.DB.FetchAndStoreNotifications(ctx, interval)
	if err != nil {
		return nil, err
	}
	fmt.Println("Notification Group: ", notificationgroup, "|| Interval: ", interval)
	notifications := []*Notification{}
	for _, row := range notificationgroup {
		var notification Notification
		notification.ID = int64(row.FeedID.ID())
		notification.Feed_ID = row.FeedID
		notification.Feed_Name = row.FeedName
		notification.Post_Count = int(row.PostCount)
		notifications = append(notifications, &notification)
	}
	return notifications, nil
}

// GetUserNotifications() is an endpoint function that retrieves all notifications
// for a specific user within a specified interval. This function currently works
// on an on-demand basis/poll basis. It is not a real-time notification system.
func (m *NotificationsModel) GetUserNotifications(userID int64, interval int64) (*NotificationsGroup, error) {
	// Create a new context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Get notifications for the specific user
	postNotifications, err := m.DB.GetUserNotifications(ctx, database.GetUserNotificationsParams{
		UserID:  userID,
		Column2: interval,
	})
	// If any errors are found, we return it
	if err != nil {
		return nil, err
	}
	// Get user comment nitifications
	commentNotifications, err := m.GetUserCommentNotifications(userID)
	// If any errors are found, we return it
	if err != nil {
		return nil, err
	}
	// Make a slice of notifications
	notifications := []*Notification{}
	// Loop through the notification group and append to the notifications slice
	for _, row := range postNotifications {
		var notification Notification
		notification.ID = int64(row.FeedID.ID())
		notification.Feed_ID = row.FeedID
		notification.Feed_Name = row.FeedName
		notification.Post_Count = int(row.PostCount)
		notifications = append(notifications, &notification)
	}
	// now let's aggregate the data
	var notificationsGroup NotificationsGroup
	notificationsGroup.Notification = notifications
	notificationsGroup.CommentNotification = commentNotifications
	// Return the notifications
	return &notificationsGroup, nil
}

// GetUserCommentNotifications() is an endpoint function that retrieves all comment notifications
// for a user. The function takes in a user id and returns a slice of CommentNotification structs.
// the notifications returned can either be for a post that a user has FAVORITED or replies
// from other users on a comment that the user has made on any post..
func (m *NotificationsModel) GetUserCommentNotifications(userID int64) ([]*CommentNotification, error) {
	// Create a new context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Get all comment notifications
	commentnotificationgroup, err := m.DB.GetUserCommentNotifications(ctx, userID)
	// If any errors are found, we return it
	if err != nil {
		return nil, err
	}
	// Make a slice of comment notifications
	commentnotifications := []*CommentNotification{}
	// Loop through the comment notification group and append to the comment notifications slice
	for _, row := range commentnotificationgroup {
		var commentnotification CommentNotification
		commentnotification.ID = int64(row.NotificationID)
		commentnotification.Comment_ID = row.CommentID
		commentnotification.User_ID = userID
		commentnotification.Post_ID = row.PostID
		commentnotification.Created_At = row.CreatedAt
		commentnotifications = append(commentnotifications, &commentnotification)
	}
	// Return the comment notifications
	return commentnotifications, nil
}

// InsertNotifications() inserts a new notification into our notifications table.
// Uses the passed in notification struct and returns an id of the inserted notification.
func (m *NotificationsModel) InsertNotifications(notification *Notification) (int32, error) {
	// Create a new context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	notoficationID, err := m.DB.InsertNotifications(ctx, database.InsertNotificationsParams{
		FeedID:    notification.Feed_ID,
		FeedName:  notification.Feed_Name,
		PostCount: int32(notification.Post_Count),
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return 0, err
	}
	return notoficationID, nil
}

// DeleteOldNotifications deletes all notifications older than the specified interval
// This should be an automated cleanup task running each 100mins or so.
func (m *NotificationsModel) ClearOldNotifications(interval int64) error {
	// Create a new context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := m.DB.ClearNotifications(ctx, interval)
	if err != nil {
		return err
	}
	return nil
}

// DeleteReadCommentNotification() deletes a read comment notification from the database
func (m *NotificationsModel) DeleteReadCommentNotification(notificationID int64) error {
	// Create a new context with a 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Delete the notification from the database
	err := m.DB.DeleteReadCommentNotification(ctx, int32(notificationID))
	if err != nil {
		return err
	}
	return nil
}
