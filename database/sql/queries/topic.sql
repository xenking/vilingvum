-- name: CreateTopic :one
INSERT INTO topics (type, content)
VALUES ($1, $2)
RETURNING id;

-- name: UpdateTopicRelation :exec
UPDATE topics
SET next_topic_id = $2
WHERE id = $1;

-- name: GetTopic :one
SELECT id, next_topic_id, type, content
FROM topics
WHERE id = $1;

-- name: GetPreviousTopics :many
SELECT t.id, t.next_topic_id, t.type, t.content
FROM topics t
         JOIN user_answers a ON t.id = a.topic_id
WHERE a.user_id = $1
  AND t.type = $2
ORDER BY t.id DESC;