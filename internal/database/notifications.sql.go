// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: notifications.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const clearNotifications = `-- name: ClearNotifications :exec
DELETE FROM notifications
WHERE created_at <= now() - ($1 * INTERVAL '1 minute')
`

func (q *Queries) ClearNotifications(ctx context.Context, dollar_1 interface{}) error {
	_, err := q.db.ExecContext(ctx, clearNotifications, dollar_1)
	return err
}

const fetchAndStoreNotifications = `-- name: FetchAndStoreNotifications :many
SELECT
    f.id AS feed_id,
    f.name AS feed_name,
    COUNT(p.id) AS post_count
FROM
    rssfeed_posts p
INNER JOIN
    feeds f ON p.feed_id = f.id
WHERE
    p.created_at >= timezone('UTC', now()) - ($1 * INTERVAL '1 minute') AND
    p.created_at <= timezone('UTC', now())
GROUP BY
    f.id, f.name
`

type FetchAndStoreNotificationsRow struct {
	FeedID    uuid.UUID
	FeedName  string
	PostCount int64
}

func (q *Queries) FetchAndStoreNotifications(ctx context.Context, dollar_1 interface{}) ([]FetchAndStoreNotificationsRow, error) {
	rows, err := q.db.QueryContext(ctx, fetchAndStoreNotifications, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FetchAndStoreNotificationsRow
	for rows.Next() {
		var i FetchAndStoreNotificationsRow
		if err := rows.Scan(&i.FeedID, &i.FeedName, &i.PostCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserNotifications = `-- name: GetUserNotifications :many
SELECT
    n.id AS notification_id,
    n.feed_id,
    n.feed_name,
    n.post_count,
    n.created_at
FROM
    notifications n
INNER JOIN
    feed_follows ff ON n.feed_id = ff.feed_id
WHERE
    ff.user_id = $1
    AND n.created_at >= now() - ($2 * INTERVAL '1 minute')
ORDER BY
    n.created_at DESC
`

type GetUserNotificationsParams struct {
	UserID  int64
	Column2 interface{}
}

type GetUserNotificationsRow struct {
	NotificationID int32
	FeedID         uuid.UUID
	FeedName       string
	PostCount      int32
	CreatedAt      time.Time
}

func (q *Queries) GetUserNotifications(ctx context.Context, arg GetUserNotificationsParams) ([]GetUserNotificationsRow, error) {
	rows, err := q.db.QueryContext(ctx, getUserNotifications, arg.UserID, arg.Column2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserNotificationsRow
	for rows.Next() {
		var i GetUserNotificationsRow
		if err := rows.Scan(
			&i.NotificationID,
			&i.FeedID,
			&i.FeedName,
			&i.PostCount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertNotifications = `-- name: InsertNotifications :one
INSERT INTO notifications (feed_id, feed_name, post_count, created_at)
VALUES ($1, $2, $3, $4)
RETURNING ID
`

type InsertNotificationsParams struct {
	FeedID    uuid.UUID
	FeedName  string
	PostCount int32
	CreatedAt time.Time
}

func (q *Queries) InsertNotifications(ctx context.Context, arg InsertNotificationsParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, insertNotifications,
		arg.FeedID,
		arg.FeedName,
		arg.PostCount,
		arg.CreatedAt,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}
