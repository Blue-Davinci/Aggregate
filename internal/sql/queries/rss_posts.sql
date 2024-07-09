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
SELECT count(*) OVER(), rssfeed_posts.* from rssfeed_posts
JOIN feed_follows ON rssfeed_posts.feed_id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY rssfeed_posts.itempublished_at DESC
LIMIT $2 OFFSET $3;

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
SELECT count(*) OVER(),
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
    p.feed_id
FROM 
    rssfeed_posts p
JOIN 
    postfavorites f ON p.id = f.post_id
WHERE 
    f.user_id = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;