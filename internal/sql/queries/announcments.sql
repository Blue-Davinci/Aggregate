-- name: MarkAnnouncmentAsReadByUser :one
INSERT INTO announcement_reads (user_id, announcement_id)
VALUES
  ($1, $2)
RETURNING *;

-- name: GetAnnouncmentsForUser :many
SELECT a.*
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id AND ar.user_id = $1
WHERE a.is_active = TRUE
  AND ar.id IS NULL;
