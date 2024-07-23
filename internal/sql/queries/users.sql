-- name: CreateUser :one
INSERT INTO users (name, email, password_hash, activated)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version;

-- name: GetUserByEmail :one
SELECT id, created_at, name, email, password_hash, activated, version, user_img
FROM users WHERE email = $1;

-- name: Update :one
UPDATE users
SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1, user_img=$5
WHERE id = $6 AND version = $7
RETURNING version;