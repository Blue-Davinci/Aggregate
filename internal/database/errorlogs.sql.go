// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: errorlogs.sql

package database

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createScraperErrorLog = `-- name: CreateScraperErrorLog :one
INSERT INTO scraper_error_logs (
    error_type, message, feed_id, status_code, retry_attempts, admin_notified, resolved, resolution_notes, occurred_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
)
ON CONFLICT (error_type, feed_id)
DO UPDATE SET
    occurrence_count = scraper_error_logs.occurrence_count + 1,
    last_occurrence = NOW(),
    retry_attempts = scraper_error_logs.retry_attempts + 1,
    updated_at = NOW(),
    message = EXCLUDED.message, -- Optionally update message
    status_code = EXCLUDED.status_code -- Optionally update status code
RETURNING id, created_at, updated_at, occurrence_count, last_occurrence
`

type CreateScraperErrorLogParams struct {
	ErrorType       string
	Message         sql.NullString
	FeedID          uuid.UUID
	StatusCode      sql.NullInt32
	RetryAttempts   sql.NullInt32
	AdminNotified   sql.NullBool
	Resolved        sql.NullBool
	ResolutionNotes sql.NullString
	OccurredAt      sql.NullTime
}

type CreateScraperErrorLogRow struct {
	ID              int32
	CreatedAt       sql.NullTime
	UpdatedAt       sql.NullTime
	OccurrenceCount sql.NullInt32
	LastOccurrence  sql.NullTime
}

func (q *Queries) CreateScraperErrorLog(ctx context.Context, arg CreateScraperErrorLogParams) (CreateScraperErrorLogRow, error) {
	row := q.db.QueryRowContext(ctx, createScraperErrorLog,
		arg.ErrorType,
		arg.Message,
		arg.FeedID,
		arg.StatusCode,
		arg.RetryAttempts,
		arg.AdminNotified,
		arg.Resolved,
		arg.ResolutionNotes,
		arg.OccurredAt,
	)
	var i CreateScraperErrorLogRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.OccurrenceCount,
		&i.LastOccurrence,
	)
	return i, err
}

const deleteScraperErrorLogByID = `-- name: DeleteScraperErrorLogByID :one
DELETE FROM scraper_error_logs
WHERE 
    id = $1
RETURNING id
`

func (q *Queries) DeleteScraperErrorLogByID(ctx context.Context, id int32) (int32, error) {
	row := q.db.QueryRowContext(ctx, deleteScraperErrorLogByID, id)
	err := row.Scan(&id)
	return id, err
}

const getAllScraperErrorLogs = `-- name: GetAllScraperErrorLogs :many
SELECT 
    COUNT(*) OVER() AS total_count,
    sel.id, sel.error_type, sel.message, sel.feed_id, f.name as feed_name, f.url as feed_url, f.img_url as feed_img_url,
    sel.occurred_at, sel.status_code, sel.retry_attempts, sel.admin_notified, sel.resolved, 
    sel.resolution_notes, sel.created_at, sel.updated_at, sel.occurrence_count, sel.last_occurrence
FROM 
    scraper_error_logs sel
JOIN 
    feeds f ON sel.feed_id = f.id
ORDER BY 
    sel.occurrence_count DESC
LIMIT $1 OFFSET $2
`

type GetAllScraperErrorLogsParams struct {
	Limit  int32
	Offset int32
}

type GetAllScraperErrorLogsRow struct {
	TotalCount      int64
	ID              int32
	ErrorType       string
	Message         sql.NullString
	FeedID          uuid.UUID
	FeedName        string
	FeedUrl         string
	FeedImgUrl      string
	OccurredAt      sql.NullTime
	StatusCode      sql.NullInt32
	RetryAttempts   sql.NullInt32
	AdminNotified   sql.NullBool
	Resolved        sql.NullBool
	ResolutionNotes sql.NullString
	CreatedAt       sql.NullTime
	UpdatedAt       sql.NullTime
	OccurrenceCount sql.NullInt32
	LastOccurrence  sql.NullTime
}

func (q *Queries) GetAllScraperErrorLogs(ctx context.Context, arg GetAllScraperErrorLogsParams) ([]GetAllScraperErrorLogsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllScraperErrorLogs, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllScraperErrorLogsRow
	for rows.Next() {
		var i GetAllScraperErrorLogsRow
		if err := rows.Scan(
			&i.TotalCount,
			&i.ID,
			&i.ErrorType,
			&i.Message,
			&i.FeedID,
			&i.FeedName,
			&i.FeedUrl,
			&i.FeedImgUrl,
			&i.OccurredAt,
			&i.StatusCode,
			&i.RetryAttempts,
			&i.AdminNotified,
			&i.Resolved,
			&i.ResolutionNotes,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.OccurrenceCount,
			&i.LastOccurrence,
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

const getScraperErrorLogByID = `-- name: GetScraperErrorLogByID :one
SELECT 
    sel.id, sel.error_type, sel.message, sel.feed_id, f.name as feed_name, f.url as feed_url, f.img_url as feed_img_url, 
    sel.occurred_at, sel.status_code, sel.retry_attempts, sel.admin_notified, sel.resolved, 
    sel.resolution_notes, sel.created_at, sel.updated_at, sel.occurrence_count, sel.last_occurrence
FROM 
    scraper_error_logs sel
JOIN 
    feeds f ON sel.feed_id = f.id
WHERE 
    sel.id = $1
`

type GetScraperErrorLogByIDRow struct {
	ID              int32
	ErrorType       string
	Message         sql.NullString
	FeedID          uuid.UUID
	FeedName        string
	FeedUrl         string
	FeedImgUrl      string
	OccurredAt      sql.NullTime
	StatusCode      sql.NullInt32
	RetryAttempts   sql.NullInt32
	AdminNotified   sql.NullBool
	Resolved        sql.NullBool
	ResolutionNotes sql.NullString
	CreatedAt       sql.NullTime
	UpdatedAt       sql.NullTime
	OccurrenceCount sql.NullInt32
	LastOccurrence  sql.NullTime
}

func (q *Queries) GetScraperErrorLogByID(ctx context.Context, id int32) (GetScraperErrorLogByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getScraperErrorLogByID, id)
	var i GetScraperErrorLogByIDRow
	err := row.Scan(
		&i.ID,
		&i.ErrorType,
		&i.Message,
		&i.FeedID,
		&i.FeedName,
		&i.FeedUrl,
		&i.FeedImgUrl,
		&i.OccurredAt,
		&i.StatusCode,
		&i.RetryAttempts,
		&i.AdminNotified,
		&i.Resolved,
		&i.ResolutionNotes,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.OccurrenceCount,
		&i.LastOccurrence,
	)
	return i, err
}

const updateScraperErrorLog = `-- name: UpdateScraperErrorLog :one
UPDATE scraper_error_logs
SET
    admin_notified = $1,
    resolved = $2,
    resolution_notes = $3,
    updated_at = NOW()
WHERE id = $4
RETURNING id, admin_notified, resolved, resolution_notes, updated_at
`

type UpdateScraperErrorLogParams struct {
	AdminNotified   sql.NullBool
	Resolved        sql.NullBool
	ResolutionNotes sql.NullString
	ID              int32
}

type UpdateScraperErrorLogRow struct {
	ID              int32
	AdminNotified   sql.NullBool
	Resolved        sql.NullBool
	ResolutionNotes sql.NullString
	UpdatedAt       sql.NullTime
}

func (q *Queries) UpdateScraperErrorLog(ctx context.Context, arg UpdateScraperErrorLogParams) (UpdateScraperErrorLogRow, error) {
	row := q.db.QueryRowContext(ctx, updateScraperErrorLog,
		arg.AdminNotified,
		arg.Resolved,
		arg.ResolutionNotes,
		arg.ID,
	)
	var i UpdateScraperErrorLogRow
	err := row.Scan(
		&i.ID,
		&i.AdminNotified,
		&i.Resolved,
		&i.ResolutionNotes,
		&i.UpdatedAt,
	)
	return i, err
}
