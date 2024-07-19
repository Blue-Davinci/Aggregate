-- name: FetchAndStoreNotifications :many
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
    f.id, f.name;

-- name: GetUserNotifications :many
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
    n.created_at DESC;

-- name: InsertNotifications :one
INSERT INTO notifications (feed_id, feed_name, post_count, created_at)
VALUES ($1, $2, $3, $4)
RETURNING ID;


-- name: ClearNotifications :exec
DELETE FROM notifications
WHERE created_at <= now() - ($1 * INTERVAL '1 minute');

-- name: CreateCommentNotification :one
INSERT INTO comment_notifications (user_id, post_id, comment_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserCommentNotifications :many
(
    -- Comment Notifications for Favorited Posts
    SELECT
        cn.id AS notification_id,
        cn.post_id AS post_id,
        'Comment on Favorited Post' AS notification_type,
        c.id AS comment_id,
        COALESCE(c.parent_comment_id, '00000000-0000-0000-0000-000000000000') AS replied_comment_id,
        cn.created_at
    FROM
        comment_notifications cn
    INNER JOIN
        comments c ON cn.comment_id = c.id
    WHERE
        c.post_id IN (
            SELECT pf.post_id
            FROM postfavorites pf
            WHERE pf.user_id = $1
        )
)
UNION ALL
(
    -- Reply Notifications
    SELECT
        cn.id AS notification_id,
        cn.post_id AS post_id,
        'Reply to Your Comment' AS notification_type,
        c.id AS comment_id,
        COALESCE(c.parent_comment_id, '00000000-0000-0000-0000-000000000000') AS replied_comment_id,
        cn.created_at
    FROM
        comment_notifications cn
    INNER JOIN
        comments c ON cn.comment_id = c.id
    WHERE
        c.parent_comment_id IN (
            SELECT id
            FROM comments
            WHERE user_id = $1
        )
)
ORDER BY
    created_at DESC;

-- name: DeleteReadCommentNotification :exec
DELETE FROM comment_notifications
WHERE id=$1;