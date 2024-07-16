-- name: CreateRssFeedPost :one
INSERT INTO rssfeed_posts (
    id, 
    created_at, 
    updated_at, 
    channeltitle, 
    channelurl,
    channeldescription,
    channellanguage,
    itemtitle,
    itemdescription, 
    itempublished_at, 
    itemurl, 
    img_url, 
    feed_id
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9, $10, $11, $12, $13)
RETURNING *;

-- name: GetFollowedRssPostsForUser :many
SELECT 
    p.*, 
    COALESCE(pf.is_favorite, false) AS is_favorite,
    COUNT(*) OVER() AS total_count
FROM 
    rssfeed_posts p
JOIN (
    SELECT 
        ff.feed_id 
    FROM 
        feed_follows ff
    WHERE 
        ff.user_id = $1  -- Parameter 1: user_id
) ff ON p.feed_id = ff.feed_id
LEFT JOIN (
    SELECT 
        pf.post_id,
        true AS is_favorite
    FROM 
        postfavorites pf
    WHERE 
        pf.user_id = $1  -- Parameter 1: user_id
) pf ON p.id = pf.post_id
WHERE 
    ($2 = '' OR to_tsvector('simple', p.itemtitle) @@ plainto_tsquery('simple', $2))  -- Parameter 2: itemtitle (full-text search for item title)
    AND ($3::uuid = '00000000-0000-0000-0000-000000000000' OR p.feed_id = $3::uuid)  -- Parameter 3: feed_id (filter by feed_id if provided)
ORDER BY 
    p.created_at DESC
LIMIT $4 OFFSET $5;  -- Parameters 4 and 5: limit and offset


-- name: GetRSSFavoritePostsForUser :many
SELECT id, post_id, feed_id, user_id, created_at
FROM postfavorites
WHERE user_id = $1;

-- name: CreateRSSFavoritePost :one
INSERT INTO postfavorites (post_id, feed_id, user_id, created_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteRSSFavoritePost :exec
DELETE FROM postfavorites
WHERE post_id = $1 AND user_id = $2;

-- name: GetRSSFavoritePostsOnlyForUser :many
SELECT 
    COUNT(*) OVER() AS total_count,
    p.id,
    p.created_at,
    p.updated_at,
    p.channeltitle,
    p.channelurl,
    p.channeldescription,
    p.channellanguage,
    p.itemtitle,
    p.itemdescription,
    p.itempublished_at,
    p.itemurl,
    p.img_url,
    p.feed_id,
    true AS is_favorite  -- Initialize is_favorite to true
FROM 
    rssfeed_posts p
JOIN 
    postfavorites f ON p.id = f.post_id
WHERE 
    f.user_id = $1  -- Parameter 1: user_id
    AND ($2 = '' OR to_tsvector('simple', p.itemtitle) @@ plainto_tsquery('simple', $2))  -- Parameter 2: itemtitle (full-text search for item title)
    AND ($3::uuid = '00000000-0000-0000-0000-000000000000' OR p.feed_id = $3::uuid)  -- Parameter 3: feed_id (filter by feed_id if provided)
ORDER BY 
    p.created_at DESC
LIMIT $4 OFFSET $5;  -- Parameters 4 and 5: limit and offset
