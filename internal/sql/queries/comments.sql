-- name: CreateComments :one
INSERT INTO comments (id, post_id, user_id, parent_comment_id, comment_text)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCommentsForPost :many
SELECT 
    comments.id, 
    comments.post_id, 
    comments.user_id, 
    users.name as user_name, 
    comments.parent_comment_id, 
    comments.comment_text, 
    comments.created_at
FROM comments
JOIN users ON comments.user_id = users.id
WHERE comments.post_id = $1;

