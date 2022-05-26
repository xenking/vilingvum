-- name: CreateTopic :one
INSERT INTO topics (type, content)
VALUES ($1, $2)
RETURNING id;

-- name: GetTopic :one
SELECT id, next_topic_id, type, content
FROM topics
WHERE id = $1;

-- name: GetLastTopicID :one
SELECT topic_id
FROM user_answers
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: InsertUserAnswer :exec
INSERT INTO user_answers (user_id, topic_id, response)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, topic_id) DO UPDATE SET response   = $3,
                                              updated_at = NOW();

-- name: GetPreviousTopics :many
SELECT t.id, t.content
FROM topics t
         JOIN user_answers a ON t.id = a.topic_id
WHERE a.user_id = $1
  AND t.type = $2
ORDER BY t.id DESC;