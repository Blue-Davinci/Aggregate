-- name: GetUserLimitations :one
WITH user_activities AS (
  SELECT
    u.id AS user_id, 
    COUNT(DISTINCT ff.feed_id) AS followed_feeds,
    COUNT(DISTINCT f.id) AS created_feeds,
    COUNT(DISTINCT c.id) AS comments_today
  FROM
    users u
  LEFT JOIN
    feed_follows ff ON ff.user_id = u.id
  LEFT JOIN
    feeds f ON f.user_id = u.id
  LEFT JOIN
    comments c ON c.user_id = u.id AND DATE(c.created_at) = CURRENT_DATE
  WHERE
    u.id = $1
  GROUP BY
    u.id
)

SELECT * FROM user_activities;
