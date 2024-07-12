// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: feeds.sql

package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createFeed = `-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id, img_url, feed_type, feed_description) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
RETURNING id, created_at, updated_at, name, url, version, user_id, img_url, last_fetched_at, feed_type, feed_description
`

type CreateFeedParams struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Url             string
	UserID          int64
	ImgUrl          string
	FeedType        string
	FeedDescription string
}

func (q *Queries) CreateFeed(ctx context.Context, arg CreateFeedParams) (Feed, error) {
	row := q.db.QueryRowContext(ctx, createFeed,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.Name,
		arg.Url,
		arg.UserID,
		arg.ImgUrl,
		arg.FeedType,
		arg.FeedDescription,
	)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.Version,
		&i.UserID,
		&i.ImgUrl,
		&i.LastFetchedAt,
		&i.FeedType,
		&i.FeedDescription,
	)
	return i, err
}

const createFeedFollow = `-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, created_at, updated_at, user_id, feed_id
`

type CreateFeedFollowParams struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    int64
	FeedID    uuid.UUID
}

func (q *Queries) CreateFeedFollow(ctx context.Context, arg CreateFeedFollowParams) (FeedFollow, error) {
	row := q.db.QueryRowContext(ctx, createFeedFollow,
		arg.ID,
		arg.CreatedAt,
		arg.UpdatedAt,
		arg.UserID,
		arg.FeedID,
	)
	var i FeedFollow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.FeedID,
	)
	return i, err
}

const deleteFeedFollow = `-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE id = $1 
AND user_id = $2
`

type DeleteFeedFollowParams struct {
	ID     uuid.UUID
	UserID int64
}

func (q *Queries) DeleteFeedFollow(ctx context.Context, arg DeleteFeedFollowParams) error {
	_, err := q.db.ExecContext(ctx, deleteFeedFollow, arg.ID, arg.UserID)
	return err
}

const getAllFeeds = `-- name: GetAllFeeds :many
SELECT count(*) OVER(), id, created_at, updated_at, name, url, user_id, version, img_url, feed_type, feed_description
FROM feeds
WHERE ($1 = '' OR to_tsvector('simple', name) @@ plainto_tsquery('simple', $1))
AND ($2 = '' OR url LIKE '%' || $2 || '%')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4
`

type GetAllFeedsParams struct {
	Column1 interface{}
	Column2 interface{}
	Limit   int32
	Offset  int32
}

type GetAllFeedsRow struct {
	Count           int64
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Url             string
	UserID          int64
	Version         int32
	ImgUrl          string
	FeedType        string
	FeedDescription string
}

func (q *Queries) GetAllFeeds(ctx context.Context, arg GetAllFeedsParams) ([]GetAllFeedsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllFeeds,
		arg.Column1,
		arg.Column2,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllFeedsRow
	for rows.Next() {
		var i GetAllFeedsRow
		if err := rows.Scan(
			&i.Count,
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.UserID,
			&i.Version,
			&i.ImgUrl,
			&i.FeedType,
			&i.FeedDescription,
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

const getAllFeedsFollowedByUser = `-- name: GetAllFeedsFollowedByUser :many
SELECT 
    f.id, 
    f.created_at, 
    f.updated_at, 
    f.name, 
    f.url, 
    f.version, 
    f.user_id, 
    f.img_url, 
    f.last_fetched_at, 
    f.feed_type, 
    f.feed_description, 
    COALESCE(ff.is_followed, false) AS is_followed,
    COUNT(*) OVER() AS follow_count
FROM 
    feeds f
LEFT JOIN (
    SELECT 
        feed_id, 
        true AS is_followed 
    FROM 
        feed_follows 
    WHERE 
        feed_follows.user_id = $1
) ff ON f.id = ff.feed_id
WHERE 
    (to_tsvector('simple', f.name) @@ plainto_tsquery('simple', $2) OR $2 = '')
ORDER BY 
    f.created_at DESC
LIMIT $3 OFFSET $4
`

type GetAllFeedsFollowedByUserParams struct {
	UserID         int64
	PlaintoTsquery string
	Limit          int32
	Offset         int32
}

type GetAllFeedsFollowedByUserRow struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Url             string
	Version         int32
	UserID          int64
	ImgUrl          string
	LastFetchedAt   sql.NullTime
	FeedType        string
	FeedDescription string
	IsFollowed      bool
	FollowCount     int64
}

func (q *Queries) GetAllFeedsFollowedByUser(ctx context.Context, arg GetAllFeedsFollowedByUserParams) ([]GetAllFeedsFollowedByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllFeedsFollowedByUser,
		arg.UserID,
		arg.PlaintoTsquery,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllFeedsFollowedByUserRow
	for rows.Next() {
		var i GetAllFeedsFollowedByUserRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.Version,
			&i.UserID,
			&i.ImgUrl,
			&i.LastFetchedAt,
			&i.FeedType,
			&i.FeedDescription,
			&i.IsFollowed,
			&i.FollowCount,
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

const getNextFeedsToFetch = `-- name: GetNextFeedsToFetch :many
SELECT id, created_at, updated_at, name, url, version, user_id, img_url, last_fetched_at, feed_type, feed_description FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1
`

func (q *Queries) GetNextFeedsToFetch(ctx context.Context, limit int32) ([]Feed, error) {
	rows, err := q.db.QueryContext(ctx, getNextFeedsToFetch, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Feed
	for rows.Next() {
		var i Feed
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.Version,
			&i.UserID,
			&i.ImgUrl,
			&i.LastFetchedAt,
			&i.FeedType,
			&i.FeedDescription,
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

const getTopFollowedFeeds = `-- name: GetTopFollowedFeeds :many
SELECT f.id, f.created_at, f.updated_at, f.name, f.url, f.version, f.user_id, f.img_url, f.last_fetched_at, f.feed_type, f.feed_description, ff.follow_count
FROM (
    SELECT feed_id, COUNT(*) AS follow_count
    FROM feed_follows
    GROUP BY feed_id
    ORDER BY follow_count DESC
    LIMIT $1
) AS ff
JOIN feeds f ON f.id = ff.feed_id
ORDER BY ff.follow_count DESC
`

type GetTopFollowedFeedsRow struct {
	ID              uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Name            string
	Url             string
	Version         int32
	UserID          int64
	ImgUrl          string
	LastFetchedAt   sql.NullTime
	FeedType        string
	FeedDescription string
	FollowCount     int64
}

func (q *Queries) GetTopFollowedFeeds(ctx context.Context, limit int32) ([]GetTopFollowedFeedsRow, error) {
	rows, err := q.db.QueryContext(ctx, getTopFollowedFeeds, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTopFollowedFeedsRow
	for rows.Next() {
		var i GetTopFollowedFeedsRow
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Name,
			&i.Url,
			&i.Version,
			&i.UserID,
			&i.ImgUrl,
			&i.LastFetchedAt,
			&i.FeedType,
			&i.FeedDescription,
			&i.FollowCount,
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

const markFeedAsFetched = `-- name: MarkFeedAsFetched :one
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at, name, url, version, user_id, img_url, last_fetched_at, feed_type, feed_description
`

func (q *Queries) MarkFeedAsFetched(ctx context.Context, id uuid.UUID) (Feed, error) {
	row := q.db.QueryRowContext(ctx, markFeedAsFetched, id)
	var i Feed
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Name,
		&i.Url,
		&i.Version,
		&i.UserID,
		&i.ImgUrl,
		&i.LastFetchedAt,
		&i.FeedType,
		&i.FeedDescription,
	)
	return i, err
}
