-- name: CreateComments :one
INSERT INTO comments (id, post_id, user_id, parent_comment_id, comment_text)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCommentsForPost :many
SELECT id, post_id, user_id, parent_comment_id, comment_text, created_at
FROM comments
WHERE post_id = $1;
