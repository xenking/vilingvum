-- name: CreateUser :one
INSERT INTO users (id, name, username, invite_code, state, active_until)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1;

-- name: IsUserExists :one
SELECT true
FROM users
WHERE id = $1
LIMIT 1;

-- name: ListActiveUsers :many
SELECT id, name, settings, is_admin
FROM users
WHERE active_until > now()
  AND state = 'active';

-- name: ListAdmins :many
SELECT id
FROM users
WHERE is_admin = true;
