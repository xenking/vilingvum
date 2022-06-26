-- name: AddDictionaryRecord :exec
INSERT INTO dictionary(topic_id, word, meaning)
VALUES ($1, $2, $3);

-- name: GetDictionary :many
SELECT id, topic_id, word, meaning
FROM dictionary
ORDER BY topic_id;