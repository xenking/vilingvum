-- name: AddDictionaryRecord :exec
INSERT INTO dictionary(topic_id, word, meaning)
VALUES ($1, $2, $3);

-- name: GetDictionary :many
SELECT word, meaning
FROM dictionary
WHERE topic_id <= $1
ORDER BY id;