-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id, email, role;

-- name: UpdateUserName :exec
UPDATE users SET name=$1 WHERE id=$2;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash=$1 WHERE id=$2;

-- name: GetUserByID :one
SELECT id, name, email, role, created_at, updated_at FROM users WHERE id=$1;

-- name: GetUserByEmail :one
SELECT id, name, email, role, created_at, updated_at FROM users WHERE email=$1;