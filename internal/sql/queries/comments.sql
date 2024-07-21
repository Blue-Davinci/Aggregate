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
    comments.created_at,
    comments.version,
    CASE WHEN comments.user_id = $2 THEN true ELSE false END AS isEditable
FROM comments
JOIN users ON comments.user_id = users.id
WHERE comments.post_id = $1;

-- name: UpdateUserComment :one
UPDATE comments
SET comment_text = $1, updated_at = now(), version = version + 1
WHERE id = $2 AND user_id = $3 AND version = $4
RETURNING version;

-- name: GetCommentByID :one
SELECT 
    id,
    post_id,
    user_id,
    parent_comment_id,
    comment_text,
    created_at,
    updated_at,
    version
FROM comments
WHERE id = $1 AND user_id = $2;
