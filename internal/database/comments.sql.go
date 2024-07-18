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
RETURNING id, post_id, user_id, parent_comment_id, comment_text, created_at, updated_at
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
	)
	return i, err
}

const getCommentsForPost = `-- name: GetCommentsForPost :many
SELECT id, post_id, user_id, parent_comment_id, comment_text, created_at
FROM comments
WHERE post_id = $1
`

type GetCommentsForPostRow struct {
	ID              uuid.UUID
	PostID          uuid.UUID
	UserID          int64
	ParentCommentID uuid.NullUUID
	CommentText     string
	CreatedAt       time.Time
}

func (q *Queries) GetCommentsForPost(ctx context.Context, postID uuid.UUID) ([]GetCommentsForPostRow, error) {
	rows, err := q.db.QueryContext(ctx, getCommentsForPost, postID)
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
			&i.ParentCommentID,
			&i.CommentText,
			&i.CreatedAt,
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
