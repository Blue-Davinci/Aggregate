-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id, img_url, feed_type, feed_description) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
RETURNING *;

-- name: GetAllFeeds :many
SELECT count(*) OVER(), id, created_at, updated_at, name, url, user_id, version, img_url, feed_type, feed_description
FROM feeds
WHERE ($1 = '' OR to_tsvector('simple', name) @@ plainto_tsquery('simple', $1))
AND ($2 = '' OR url LIKE '%' || $2 || '%')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAllFeedsFollowedByUser :many
SELECT count(*) OVER(), id, created_at, updated_at, user_id, feed_id
FROM feed_follows
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE id = $1 
AND user_id = $2;

-- name: GetNextFeedsToFetch :many
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1;

-- name: MarkFeedAsFetched :one
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;