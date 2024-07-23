-- name: InsertApiKey :one
INSERT INTO api_keys (api_key, user_id, expiry, scope)
VALUES ($1, $2, $3, $4)
RETURNING user_id;

-- name: DeletAllAPIKeysForUser :exec
DELETE FROM api_keys
WHERE scope = $1 AND user_id = $2;

-- name: GetForToken :one
SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version, users.user_img
FROM users
INNER JOIN api_keys
ON users.id = api_keys.user_id
WHERE api_keys.api_key = $1
AND api_keys.scope = $2
AND api_keys.expiry > $3;