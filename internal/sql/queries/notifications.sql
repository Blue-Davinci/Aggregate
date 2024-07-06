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
WHERE created_at <= now() - ($1 * INTERVAL '90 minute');