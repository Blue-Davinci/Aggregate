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

const adminGetFeedsPendingApproval = `-- name: AdminGetFeedsPendingApproval :many

SELECT 
    count(*) OVER() AS total_count,
    feeds.id, 
    feeds.created_at, 
    feeds.updated_at, 
    feeds.name, 
    feeds.url, 
    feeds.user_id, 
    feeds.version, 
    feeds.img_url, 
    feeds.feed_type, 
    feeds.feed_description, 
    feeds.is_hidden, 
    feeds.approval_status, 
    feeds.priority,
    users.id AS user_id,
    users.name AS user_name,
    users.user_img AS user_img
FROM 
    feeds
JOIN 
    users 
ON 
    feeds.user_id = users.id
WHERE 
    ($1 = '' OR to_tsvector('simple', feeds.name) @@ plainto_tsquery('simple', $1))
    AND (feeds.feed_type = $2 OR $2 = '')
    AND feeds.approval_status = 'pending'
ORDER BY 
    feeds.created_at DESC
LIMIT 
    $3 OFFSET $4
`

type AdminGetFeedsPendingApprovalParams struct {
	Column1  interface{}
	FeedType string
	Limit    int32
	Offset   int32
}

type AdminGetFeedsPendingApprovalRow struct {
	TotalCount      int64
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
	IsHidden        bool
	ApprovalStatus  string
	Priority        string
	UserID_2        int64
	UserName        string
	UserImg         string
}

//---------------------------------------------------------------------------
// ADMIN
//---------------------------------------------------------------------------
func (q *Queries) AdminGetFeedsPendingApproval(ctx context.Context, arg AdminGetFeedsPendingApprovalParams) ([]AdminGetFeedsPendingApprovalRow, error) {
	rows, err := q.db.QueryContext(ctx, adminGetFeedsPendingApproval,
		arg.Column1,
		arg.FeedType,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AdminGetFeedsPendingApprovalRow
	for rows.Next() {
		var i AdminGetFeedsPendingApprovalRow
		if err := rows.Scan(
			&i.TotalCount,
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
			&i.IsHidden,
			&i.ApprovalStatus,
			&i.Priority,
			&i.UserID_2,
			&i.UserName,
			&i.UserImg,
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

const createFeed = `-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id, img_url, feed_type, feed_description, is_hidden) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
RETURNING id, created_at, updated_at, name, url, version, user_id, img_url, last_fetched_at, feed_type, feed_description, is_hidden, approval_status, priority
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
	IsHidden        bool
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
		arg.IsHidden,
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
		&i.IsHidden,
		&i.ApprovalStatus,
		&i.Priority,
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
SELECT count(*) OVER(), id, created_at, updated_at, name, url, user_id, version, img_url, feed_type, feed_description, is_hidden
FROM feeds
WHERE ($1 = '' OR to_tsvector('simple', name) @@ plainto_tsquery('simple', $1))
AND feed_type = $2 OR $2 = ''
AND is_hidden = FALSE
AND approval_status='approved'
ORDER BY created_at DESC
LIMIT $3 OFFSET $4
`

type GetAllFeedsParams struct {
	Column1  interface{}
	FeedType string
	Limit    int32
	Offset   int32
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
	IsHidden        bool
}

func (q *Queries) GetAllFeeds(ctx context.Context, arg GetAllFeedsParams) ([]GetAllFeedsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllFeeds,
		arg.Column1,
		arg.FeedType,
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
			&i.IsHidden,
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
    f.is_hidden,
    COALESCE(ff.is_followed, false) AS is_followed,
    ff.follow_id,
    COUNT(*) OVER() AS follow_count
FROM 
    feeds f
LEFT JOIN (
    SELECT 
        feed_id, 
        id AS follow_id,
        true AS is_followed 
    FROM 
        feed_follows 
    WHERE 
        feed_follows.user_id = $1
) ff ON f.id = ff.feed_id
WHERE 
    (to_tsvector('simple', f.name) @@ plainto_tsquery('simple', $2) OR $2 = '')
    AND (f.is_hidden = false OR f.user_id = $1)
    AND f.approval_status = 'approved'
    AND (f.feed_type = $5 OR $5 = '')
ORDER BY 
    f.created_at DESC
LIMIT $3 OFFSET $4
`

type GetAllFeedsFollowedByUserParams struct {
	UserID         int64
	PlaintoTsquery string
	Limit          int32
	Offset         int32
	FeedType       string
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
	IsHidden        bool
	IsFollowed      bool
	FollowID        uuid.UUID
	FollowCount     int64
}

func (q *Queries) GetAllFeedsFollowedByUser(ctx context.Context, arg GetAllFeedsFollowedByUserParams) ([]GetAllFeedsFollowedByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllFeedsFollowedByUser,
		arg.UserID,
		arg.PlaintoTsquery,
		arg.Limit,
		arg.Offset,
		arg.FeedType,
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
			&i.IsHidden,
			&i.IsFollowed,
			&i.FollowID,
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

const getFeedById = `-- name: GetFeedById :one
SELECT id, created_at, updated_at, name, url, user_id, version, img_url, feed_type, feed_description, is_hidden
FROM feeds
WHERE id = $1
`

type GetFeedByIdRow struct {
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
	IsHidden        bool
}

func (q *Queries) GetFeedById(ctx context.Context, id uuid.UUID) (GetFeedByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getFeedById, id)
	var i GetFeedByIdRow
	err := row.Scan(
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
		&i.IsHidden,
	)
	return i, err
}

const getFeedSearchOptions = `-- name: GetFeedSearchOptions :many
SELECT DISTINCT id, name
FROM feeds
`

type GetFeedSearchOptionsRow struct {
	ID   uuid.UUID
	Name string
}

func (q *Queries) GetFeedSearchOptions(ctx context.Context) ([]GetFeedSearchOptionsRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedSearchOptions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedSearchOptionsRow
	for rows.Next() {
		var i GetFeedSearchOptionsRow
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
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

const getFeedTypeSearchOptions = `-- name: GetFeedTypeSearchOptions :many
SELECT DISTINCT feed_type
FROM feeds
`

func (q *Queries) GetFeedTypeSearchOptions(ctx context.Context) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, getFeedTypeSearchOptions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var feed_type string
		if err := rows.Scan(&feed_type); err != nil {
			return nil, err
		}
		items = append(items, feed_type)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getFeedUserAndStatisticsByID = `-- name: GetFeedUserAndStatisticsByID :one
SELECT
    u.name AS user_name,
    u.user_img AS user_img_url,
    COUNT(pf.id) AS liked_posts_count
FROM
    feeds f
JOIN
    users u ON f.user_id = u.id
LEFT JOIN
    rssfeed_posts rp ON rp.feed_id = f.id
LEFT JOIN
    postfavorites pf ON pf.post_id = rp.id
WHERE
    f.id = $1
GROUP BY
    u.id
`

type GetFeedUserAndStatisticsByIDRow struct {
	UserName        string
	UserImgUrl      string
	LikedPostsCount int64
}

func (q *Queries) GetFeedUserAndStatisticsByID(ctx context.Context, id uuid.UUID) (GetFeedUserAndStatisticsByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getFeedUserAndStatisticsByID, id)
	var i GetFeedUserAndStatisticsByIDRow
	err := row.Scan(&i.UserName, &i.UserImgUrl, &i.LikedPostsCount)
	return i, err
}

const getFeedsCreatedByUser = `-- name: GetFeedsCreatedByUser :many
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
    f.is_hidden,
    f.approval_status,
    COALESCE(ff.follow_count, 0) AS follow_count,
    COUNT(*) OVER() AS total_count,
    fr.rejected_by,                  -- Add the rejector's username
    fr.reason AS rejection_reason,   -- Add the rejection reason
    fr.rejected_at,                  -- Add the rejection timestamp
    u.name AS rejected_by_username   -- Add the rejector's username        
FROM 
    feeds f
LEFT JOIN (
    SELECT 
        feed_id, 
        COUNT(*) AS follow_count
    FROM 
        feed_follows
    GROUP BY 
        feed_id
) ff ON f.id = ff.feed_id
LEFT JOIN 
    feed_rejections fr ON f.id = fr.feed_id AND f.approval_status = 'rejected' -- Join on feed_rejections with approval_status check
LEFT JOIN 
    users u ON fr.rejected_by = u.id  -- Join with users to get the rejector's username
WHERE 
    f.user_id = $1
    AND (to_tsvector('simple', f.name) @@ plainto_tsquery('simple', $2) OR $2 = '')
ORDER BY
    CASE 
        WHEN f.approval_status = 'pending' THEN 1
        ELSE 2
    END,
    f.created_at DESC
LIMIT $3 OFFSET $4
`

type GetFeedsCreatedByUserParams struct {
	UserID         int64
	PlaintoTsquery string
	Limit          int32
	Offset         int32
}

type GetFeedsCreatedByUserRow struct {
	ID                 uuid.UUID
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Name               string
	Url                string
	Version            int32
	UserID             int64
	ImgUrl             string
	LastFetchedAt      sql.NullTime
	FeedType           string
	FeedDescription    string
	IsHidden           bool
	ApprovalStatus     string
	FollowCount        int64
	TotalCount         int64
	RejectedBy         sql.NullInt64
	RejectionReason    sql.NullString
	RejectedAt         sql.NullTime
	RejectedByUsername sql.NullString
}

func (q *Queries) GetFeedsCreatedByUser(ctx context.Context, arg GetFeedsCreatedByUserParams) ([]GetFeedsCreatedByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getFeedsCreatedByUser,
		arg.UserID,
		arg.PlaintoTsquery,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetFeedsCreatedByUserRow
	for rows.Next() {
		var i GetFeedsCreatedByUserRow
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
			&i.IsHidden,
			&i.ApprovalStatus,
			&i.FollowCount,
			&i.TotalCount,
			&i.RejectedBy,
			&i.RejectionReason,
			&i.RejectedAt,
			&i.RejectedByUsername,
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

const getListOfFollowedFeeds = `-- name: GetListOfFollowedFeeds :many
SELECT 
    f.id, 
    f.name, 
    f.url, 
    f.feed_type, 
    f.created_at, 
    f.updated_at, 
    f.img_url,
    COUNT(*) OVER() as total_count
FROM 
    feed_follows ff
JOIN 
    feeds f ON ff.feed_id = f.id
WHERE 
    ff.user_id = $1
    AND (to_tsvector('simple', f.name) @@ plainto_tsquery('simple', $2) OR $2 = '')
ORDER BY 
    f.created_at DESC
LIMIT $3 OFFSET $4
`

type GetListOfFollowedFeedsParams struct {
	UserID         int64
	PlaintoTsquery string
	Limit          int32
	Offset         int32
}

type GetListOfFollowedFeedsRow struct {
	ID         uuid.UUID
	Name       string
	Url        string
	FeedType   string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ImgUrl     string
	TotalCount int64
}

func (q *Queries) GetListOfFollowedFeeds(ctx context.Context, arg GetListOfFollowedFeedsParams) ([]GetListOfFollowedFeedsRow, error) {
	rows, err := q.db.QueryContext(ctx, getListOfFollowedFeeds,
		arg.UserID,
		arg.PlaintoTsquery,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetListOfFollowedFeedsRow
	for rows.Next() {
		var i GetListOfFollowedFeedsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Url,
			&i.FeedType,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.ImgUrl,
			&i.TotalCount,
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
SELECT id, created_at, updated_at, name, url, version, user_id, img_url, last_fetched_at, feed_type, feed_description, is_hidden, approval_status, priority FROM feeds
WHERE approval_status = 'approved'
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
			&i.IsHidden,
			&i.ApprovalStatus,
			&i.Priority,
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

const getTopFeedCreators = `-- name: GetTopFeedCreators :many
WITH feed_follow_counts AS (
    SELECT
        f.user_id,
        COUNT(ff.id) AS total_follows
    FROM
        feeds f
    LEFT JOIN
        feed_follows ff ON f.id = ff.feed_id
    GROUP BY
        f.user_id
),
post_like_counts AS (
    SELECT
        f.user_id,
        COUNT(pf.id) AS total_likes
    FROM
        feeds f
    LEFT JOIN
        rssfeed_posts rp ON f.id = rp.feed_id
    LEFT JOIN
        postfavorites pf ON rp.id = pf.post_id
    GROUP BY
        f.user_id
),
feed_creation_times AS (
    SELECT
        f.user_id,
        f.created_at,
        LAG(f.created_at) OVER (PARTITION BY f.user_id ORDER BY f.created_at) AS prev_created_at
    FROM
        feeds f
),
avg_time_between_feeds AS (
    SELECT
        user_id,
        EXTRACT(EPOCH FROM AVG(created_at - prev_created_at)) / (60 * 60 * 24) AS avg_time_between_feeds
    FROM
        feed_creation_times
    WHERE
        prev_created_at IS NOT NULL
    GROUP BY
        user_id
),
feed_creation_counts AS (
    SELECT
        f.user_id,
        COUNT(f.id) AS total_created_feeds
    FROM
        feeds f
    WHERE 
        f.approval_status = 'approved'
    GROUP BY
        f.user_id
),
comment_counts AS (
    SELECT
        c.user_id,
        COUNT(c.id) AS total_comments
    FROM
        comments c
    GROUP BY
        c.user_id
)
SELECT
    u.name,
    u.user_img,
    ffc.total_follows,
    plc.total_likes,
    fcc.total_created_feeds,
    COALESCE(atbf.avg_time_between_feeds::float8, 0) AS avg_time_between_feeds,
    COALESCE(cc.total_comments, 0) AS total_comments
FROM
    feed_follow_counts ffc
LEFT JOIN
    post_like_counts plc ON ffc.user_id = plc.user_id
LEFT JOIN
    feed_creation_counts fcc ON ffc.user_id = fcc.user_id
LEFT JOIN
    avg_time_between_feeds atbf ON ffc.user_id = atbf.user_id
LEFT JOIN
    comment_counts cc ON ffc.user_id = cc.user_id
LEFT JOIN
    users u ON ffc.user_id = u.id
ORDER BY
    ffc.total_follows DESC
LIMIT $1 OFFSET $2
`

type GetTopFeedCreatorsParams struct {
	Limit  int32
	Offset int32
}

type GetTopFeedCreatorsRow struct {
	Name                sql.NullString
	UserImg             sql.NullString
	TotalFollows        int64
	TotalLikes          sql.NullInt64
	TotalCreatedFeeds   sql.NullInt64
	AvgTimeBetweenFeeds interface{}
	TotalComments       int64
}

func (q *Queries) GetTopFeedCreators(ctx context.Context, arg GetTopFeedCreatorsParams) ([]GetTopFeedCreatorsRow, error) {
	rows, err := q.db.QueryContext(ctx, getTopFeedCreators, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTopFeedCreatorsRow
	for rows.Next() {
		var i GetTopFeedCreatorsRow
		if err := rows.Scan(
			&i.Name,
			&i.UserImg,
			&i.TotalFollows,
			&i.TotalLikes,
			&i.TotalCreatedFeeds,
			&i.AvgTimeBetweenFeeds,
			&i.TotalComments,
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
SELECT f.id, f.created_at, f.updated_at, f.name, f.url, f.version, f.user_id, f.img_url, f.last_fetched_at, f.feed_type, f.feed_description, f.is_hidden, f.approval_status, f.priority, ff.follow_count
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
	IsHidden        bool
	ApprovalStatus  string
	Priority        string
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
			&i.IsHidden,
			&i.ApprovalStatus,
			&i.Priority,
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
RETURNING id, created_at, updated_at, name, url, version, user_id, img_url, last_fetched_at, feed_type, feed_description, is_hidden, approval_status, priority
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
		&i.IsHidden,
		&i.ApprovalStatus,
		&i.Priority,
	)
	return i, err
}

const updateFeed = `-- name: UpdateFeed :exec
UPDATE feeds
SET updated_at = NOW(), name = $2, url = $3, version = version + 1, img_url = $4, feed_type = $5, feed_description = $6, is_hidden = $7
WHERE id = $1 AND version = $8
`

type UpdateFeedParams struct {
	ID              uuid.UUID
	Name            string
	Url             string
	ImgUrl          string
	FeedType        string
	FeedDescription string
	IsHidden        bool
	Version         int32
}

func (q *Queries) UpdateFeed(ctx context.Context, arg UpdateFeedParams) error {
	_, err := q.db.ExecContext(ctx, updateFeed,
		arg.ID,
		arg.Name,
		arg.Url,
		arg.ImgUrl,
		arg.FeedType,
		arg.FeedDescription,
		arg.IsHidden,
		arg.Version,
	)
	return err
}
