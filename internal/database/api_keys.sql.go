// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: api_keys.sql

package database

import (
	"context"
	"time"
)

const deletAllAPIKeysForUser = `-- name: DeletAllAPIKeysForUser :exec
DELETE FROM api_keys
WHERE scope = $1 AND user_id = $2
`

type DeletAllAPIKeysForUserParams struct {
	Scope  string
	UserID int64
}

func (q *Queries) DeletAllAPIKeysForUser(ctx context.Context, arg DeletAllAPIKeysForUserParams) error {
	_, err := q.db.ExecContext(ctx, deletAllAPIKeysForUser, arg.Scope, arg.UserID)
	return err
}

const getForToken = `-- name: GetForToken :one
SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
FROM users
INNER JOIN api_keys
ON users.id = api_keys.user_id
WHERE api_keys.api_key = $1
AND api_keys.scope = $2
AND api_keys.expiry > $3
`

type GetForTokenParams struct {
	ApiKey []byte
	Scope  string
	Expiry time.Time
}

func (q *Queries) GetForToken(ctx context.Context, arg GetForTokenParams) (User, error) {
	row := q.db.QueryRowContext(ctx, getForToken, arg.ApiKey, arg.Scope, arg.Expiry)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.Name,
		&i.Email,
		&i.PasswordHash,
		&i.Activated,
		&i.Version,
	)
	return i, err
}

const insertApiKey = `-- name: InsertApiKey :one
INSERT INTO api_keys (api_key, user_id, expiry, scope)
VALUES ($1, $2, $3, $4)
RETURNING user_id
`

type InsertApiKeyParams struct {
	ApiKey []byte
	UserID int64
	Expiry time.Time
	Scope  string
}

func (q *Queries) InsertApiKey(ctx context.Context, arg InsertApiKeyParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, insertApiKey,
		arg.ApiKey,
		arg.UserID,
		arg.Expiry,
		arg.Scope,
	)
	var user_id int64
	err := row.Scan(&user_id)
	return user_id, err
}
