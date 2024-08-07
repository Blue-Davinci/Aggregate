// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: comments.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createComments = `-- name: CreateComments :one
INSERT INTO comments (id, post_id, user_id, parent_comment_id, comment_text)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, post_id, user_id, parent_comment_id, comment_text, created_at, updated_at, version
`

type CreateCommentsParams struct {
	ID              uuid.UUID
	PostID          uuid.UUID
	UserID          int64
	ParentCommentID uuid.NullUUID
	CommentText     string
}

func (q *Queries) CreateComments(ctx context.Context, arg CreateCommentsParams) (Comment, error) {
	row := q.db.QueryRowContext(ctx, createComments,
		arg.ID,
		arg.PostID,
		arg.UserID,
		arg.ParentCommentID,
		arg.CommentText,
	)
	var i Comment
	err := row.Scan(
		&i.ID,
		&i.PostID,
		&i.UserID,
		&i.ParentCommentID,
		&i.CommentText,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
	)
	return i, err
}

const deleteComment = `-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1 AND user_id = $2
`

type DeleteCommentParams struct {
	ID     uuid.UUID
	UserID int64
}

func (q *Queries) DeleteComment(ctx context.Context, arg DeleteCommentParams) error {
	_, err := q.db.ExecContext(ctx, deleteComment, arg.ID, arg.UserID)
	return err
}

const getCommentByID = `-- name: GetCommentByID :one
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
WHERE id = $1 AND user_id = $2
`

type GetCommentByIDParams struct {
	ID     uuid.UUID
	UserID int64
}

func (q *Queries) GetCommentByID(ctx context.Context, arg GetCommentByIDParams) (Comment, error) {
	row := q.db.QueryRowContext(ctx, getCommentByID, arg.ID, arg.UserID)
	var i Comment
	err := row.Scan(
		&i.ID,
		&i.PostID,
		&i.UserID,
		&i.ParentCommentID,
		&i.CommentText,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Version,
	)
	return i, err
}

const getCommentsForPost = `-- name: GetCommentsForPost :many
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
WHERE comments.post_id = $1
`

type GetCommentsForPostParams struct {
	PostID uuid.UUID
	UserID int64
}

type GetCommentsForPostRow struct {
	ID              uuid.UUID
	PostID          uuid.UUID
	UserID          int64
	UserName        string
	ParentCommentID uuid.NullUUID
	CommentText     string
	CreatedAt       time.Time
	Version         int32
	Iseditable      bool
}

func (q *Queries) GetCommentsForPost(ctx context.Context, arg GetCommentsForPostParams) ([]GetCommentsForPostRow, error) {
	rows, err := q.db.QueryContext(ctx, getCommentsForPost, arg.PostID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetCommentsForPostRow
	for rows.Next() {
		var i GetCommentsForPostRow
		if err := rows.Scan(
			&i.ID,
			&i.PostID,
			&i.UserID,
			&i.UserName,
			&i.ParentCommentID,
			&i.CommentText,
			&i.CreatedAt,
			&i.Version,
			&i.Iseditable,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateUserComment = `-- name: UpdateUserComment :one
UPDATE comments
SET comment_text = $1, updated_at = now(), version = version + 1
WHERE id = $2 AND user_id = $3 AND version = $4
RETURNING version
`

type UpdateUserCommentParams struct {
	CommentText string
	ID          uuid.UUID
	UserID      int64
	Version     int32
}

func (q *Queries) UpdateUserComment(ctx context.Context, arg UpdateUserCommentParams) (int32, error) {
	row := q.db.QueryRowContext(ctx, updateUserComment,
		arg.CommentText,
		arg.ID,
		arg.UserID,
		arg.Version,
	)
	var version int32
	err := row.Scan(&version)
	return version, err
}
