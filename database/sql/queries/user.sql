-- name: CreateUser :one
INSERT INTO users (id, name, username, state)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: SetActiveUser :exec
UPDATE users
SET active_until = $2
WHERE id = $1;

-- name: GetUser :one
SELECT *
FROM users
WHERE id = $1;

-- name: ListActiveUsers :many
SELECT id, name, settings, is_admin, active_until
FROM users
WHERE state = 'active';

-- name: ListPaidUsers :many
SELECT id, name, settings, is_admin, active_until
FROM users
WHERE active_until > now()
  AND state = 'active';

-- name: ListAdmins :many
SELECT id
FROM users
WHERE is_admin = true;
