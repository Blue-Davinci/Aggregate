-- name: GetFeedById :one
SELECT id, created_at, updated_at, name, url, user_id, version, img_url, feed_type, feed_description, is_hidden
FROM feeds
WHERE id = $1;

-- name: UpdateFeed :exec
UPDATE feeds
SET updated_at = NOW(), name = $2, url = $3, version = version + 1, img_url = $4, feed_type = $5, feed_description = $6, is_hidden = $7
WHERE id = $1 AND version = $8;

-- name: GetFeedSearchOptions :many
SELECT DISTINCT id, name
FROM feeds;

-- name: GetFeedTypeSearchOptions :many
SELECT DISTINCT feed_type
FROM feeds;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id, img_url, feed_type, feed_description, is_hidden) 
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) 
RETURNING *;

-- name: GetAllFeeds :many
SELECT count(*) OVER(), id, created_at, updated_at, name, url, user_id, version, img_url, feed_type, feed_description, is_hidden
FROM feeds
WHERE ($1 = '' OR to_tsvector('simple', name) @@ plainto_tsquery('simple', $1))
AND feed_type = $2 OR $2 = ''
AND is_hidden = FALSE
AND approval_status='approved'
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CreateFeedFollow :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE id = $1 
AND user_id = $2;

-- name: GetNextFeedsToFetch :many
SELECT * FROM feeds
WHERE approval_status = 'approved'
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1;

-- name: MarkFeedAsFetched :one
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetTopFollowedFeeds :many
SELECT f.*, ff.follow_count
FROM (
    SELECT feed_id, COUNT(*) AS follow_count
    FROM feed_follows
    GROUP BY feed_id
    ORDER BY follow_count DESC
    LIMIT $1
) AS ff
JOIN feeds f ON f.id = ff.feed_id
ORDER BY ff.follow_count DESC;


-- name: GetAllFeedsFollowedByUser :many
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
LIMIT $3 OFFSET $4;


-- name: GetListOfFollowedFeeds :many
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
LIMIT $3 OFFSET $4;

-- name: GetFeedsCreatedByUser :many
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
    COUNT(*) OVER() AS total_count
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
WHERE 
    f.user_id = $1
    AND (to_tsvector('simple', f.name) @@ plainto_tsquery('simple', $2) OR $2 = '')
ORDER BY 
    f.created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetFeedUserAndStatisticsByID :one
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
    u.id;

-- name: GetTopFeedCreators :many
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
LIMIT $1 OFFSET $2;
