-- name: CreateUser :one
INSERT INTO users (id, name, username, state, active_until)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1;

-- name: ListActiveUsers :many
SELECT id, name, settings, dictionary, is_admin
FROM users
WHERE active_until > now()
  AND state = 'active';

-- name: ListAdmins :many
SELECT id
FROM users
WHERE is_admin = true;
