// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: permissions.sql

package database

import (
	"context"

	"github.com/lib/pq"
)

const addPermissionsForUser = `-- name: AddPermissionsForUser :one
INSERT INTO users_permissions (user_id, permission_id)
SELECT $1, permissions.id
FROM permissions
WHERE permissions.code = ANY($2::text[])
RETURNING user_id, permission_id
`

type AddPermissionsForUserParams struct {
	UserID  int64
	Column2 []string
}

func (q *Queries) AddPermissionsForUser(ctx context.Context, arg AddPermissionsForUserParams) (UsersPermission, error) {
	row := q.db.QueryRowContext(ctx, addPermissionsForUser, arg.UserID, pq.Array(arg.Column2))
	var i UsersPermission
	err := row.Scan(&i.UserID, &i.PermissionID)
	return i, err
}

const deletePermissionsForUser = `-- name: DeletePermissionsForUser :one
DELETE FROM users_permissions
USING permissions
WHERE users_permissions.user_id = $1
AND permissions.code = $2
AND users_permissions.permission_id = permissions.id
RETURNING permission_id
`

type DeletePermissionsForUserParams struct {
	UserID int64
	Code   string
}

func (q *Queries) DeletePermissionsForUser(ctx context.Context, arg DeletePermissionsForUserParams) (int64, error) {
	row := q.db.QueryRowContext(ctx, deletePermissionsForUser, arg.UserID, arg.Code)
	var permission_id int64
	err := row.Scan(&permission_id)
	return permission_id, err
}

const getAllPermissions = `-- name: GetAllPermissions :many
SELECT id, code
FROM permissions
`

func (q *Queries) GetAllPermissions(ctx context.Context) ([]Permission, error) {
	rows, err := q.db.QueryContext(ctx, getAllPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Permission
	for rows.Next() {
		var i Permission
		if err := rows.Scan(&i.ID, &i.Code); err != nil {
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

const getAllPermissionsForUser = `-- name: GetAllPermissionsForUser :many
SELECT permissions.code
FROM permissions
INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
INNER JOIN users ON users_permissions.user_id = users.id
WHERE users.id = $1
`

func (q *Queries) GetAllPermissionsForUser(ctx context.Context, id int64) ([]string, error) {
	rows, err := q.db.QueryContext(ctx, getAllPermissionsForUser, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		items = append(items, code)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllSuperUsersWithPermissions = `-- name: GetAllSuperUsersWithPermissions :many
SELECT 
    u.id AS user_id,
    u.name,
    u.user_img,
    p.id AS permission_id,
    p.code AS permission_code
FROM 
    users u
JOIN 
    users_permissions up ON u.id = up.user_id
JOIN 
    permissions p ON up.permission_id = p.id
`

type GetAllSuperUsersWithPermissionsRow struct {
	UserID         int64
	Name           string
	UserImg        string
	PermissionID   int64
	PermissionCode string
}

func (q *Queries) GetAllSuperUsersWithPermissions(ctx context.Context) ([]GetAllSuperUsersWithPermissionsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAllSuperUsersWithPermissions)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAllSuperUsersWithPermissionsRow
	for rows.Next() {
		var i GetAllSuperUsersWithPermissionsRow
		if err := rows.Scan(
			&i.UserID,
			&i.Name,
			&i.UserImg,
			&i.PermissionID,
			&i.PermissionCode,
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
