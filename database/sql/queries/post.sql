-- name: CreatePost :one
INSERT INTO posts (content)
VALUES ($1)
RETURNING id;

-- name: GetPost :one
SELECT id, next_post_id, content
FROM posts
WHERE id = $1;

-- name: GetNextPost :one
SELECT np.id, np.content
FROM posts p
         JOIN posts np ON np.id = p.next_post_id
WHERE p.id = $1 AND p.id != np.id;

-- name: GetLastPostID :one
SELECT pe.post_id
FROM post_entries pe
         JOIN users u ON u.id = pe.user_id
WHERE u.id = $1
ORDER BY pe.created_at DESC
LIMIT 1;

-- name: CreatePostEntry :exec
INSERT INTO post_entries (user_id, post_id, status)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, post_id) DO NOTHING;

-- name: UpdatePostEntry :exec
UPDATE post_entries
SET status     = $3,
    updated_at = now()
WHERE post_id = $1
  AND user_id = $2;

