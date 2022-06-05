-- name: GetLastTopicID :one
SELECT topic_id
FROM user_answers
WHERE user_id = $1
ORDER BY topic_id DESC
LIMIT 1;

-- name: InsertUserAnswer :exec
INSERT INTO user_answers (user_id, topic_id, response)
VALUES ($1, $2, $3);
