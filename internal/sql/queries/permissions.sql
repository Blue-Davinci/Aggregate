-- name: GetAllSuperUsersWithPermissions :many
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
    permissions p ON up.permission_id = p.id;


-- name: GetAllPermissions :many
SELECT id, code
FROM permissions;

-- name: GetAllPermissionsForUser :many
SELECT permissions.code
FROM permissions
INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
INNER JOIN users ON users_permissions.user_id = users.id
WHERE users.id = $1;

-- name: AddPermissionsForUser :one
INSERT INTO users_permissions (user_id, permission_id)
SELECT $1, permissions.id
FROM permissions
WHERE permissions.code = ANY($2::text[])
RETURNING *;

-- name: DeletePermissionsForUser :one
DELETE FROM users_permissions
USING permissions
WHERE users_permissions.user_id = $1
AND permissions.code = $2
AND users_permissions.permission_id = permissions.id
RETURNING permission_id;
