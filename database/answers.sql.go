// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: answers.sql

package database

import (
	"context"

	"github.com/jackc/pgtype"
)

const getLastTopicID = `-- name: GetLastTopicID :one
SELECT topic_id
FROM user_answers
WHERE user_id = $1
ORDER BY topic_id DESC
LIMIT 1
`

func (q *Queries) GetLastTopicID(ctx context.Context, userID int64) (int64, error) {
	row := q.db.QueryRow(ctx, getLastTopicID, userID)
	var topic_id int64
	err := row.Scan(&topic_id)
	return topic_id, err
}

const insertUserAnswer = `-- name: InsertUserAnswer :exec
INSERT INTO user_answers (user_id, topic_id, response)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, topic_id) DO UPDATE SET response   = $3,
                                              updated_at = NOW()
`

type InsertUserAnswerParams struct {
	UserID   int64        `db:"user_id" json:"user_id"`
	TopicID  int64        `db:"topic_id" json:"topic_id"`
	Response pgtype.JSONB `db:"response" json:"response"`
}

func (q *Queries) InsertUserAnswer(ctx context.Context, arg *InsertUserAnswerParams) error {
	_, err := q.db.Exec(ctx, insertUserAnswer, arg.UserID, arg.TopicID, arg.Response)
	return err
}