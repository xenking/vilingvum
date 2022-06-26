-- name: CreateTopic :one
INSERT INTO topics (type, content)
VALUES ($1, $2)
RETURNING id;

-- name: UpdateTopicRelation :exec
UPDATE topics
SET next_topic_id = $2
WHERE id = $1;

-- name: GetTopics :many
SELECT id, next_topic_id, type, content
FROM topics
ORDER BY id;

-- name: GetLatestTopicID :one
SELECT topic_id
FROM user_answers
WHERE user_id = $1
ORDER BY topic_id DESC
LIMIT 1;