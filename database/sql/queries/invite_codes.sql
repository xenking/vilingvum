-- name: CreateInviteCode :exec
INSERT INTO invite_codes(code, created_by)
VALUES ($1, $2);

-- name: GetInviteCode :one
SELECT used_at
FROM invite_codes
WHERE code = $1;

-- name: ActivateInviteCode :exec
UPDATE invite_codes
SET used_by = $2,
    used_at = now()
WHERE code = $1;