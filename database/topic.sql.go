// Code generated by sqlc. DO NOT EDIT.
// source: topic.sql

package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgtype"
)

const createTopic = `-- name: CreateTopic :one
INSERT INTO topics (type, content)
VALUES ($1, $2)
RETURNING id
`

type CreateTopicParams struct {
	Type    string       `db:"type" json:"type"`
	Content pgtype.JSONB `db:"content" json:"content"`
}

func (q *Queries) CreateTopic(ctx context.Context, arg *CreateTopicParams) (int64, error) {
	row := q.db.QueryRow(ctx, createTopic, arg.Type, arg.Content)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const getPreviousTopics = `-- name: GetPreviousTopics :many
SELECT t.id, t.next_topic_id, t.type, t.content
FROM topics t
         JOIN user_answers a ON t.id = a.topic_id
WHERE a.user_id = $1
  AND t.type = $2
ORDER BY t.id DESC
`

type GetPreviousTopicsParams struct {
	UserID int64  `db:"user_id" json:"user_id"`
	Type   string `db:"type" json:"type"`
}

type GetPreviousTopicsRow struct {
	ID          int64         `db:"id" json:"id"`
	NextTopicID sql.NullInt64 `db:"next_topic_id" json:"next_topic_id"`
	Type        string        `db:"type" json:"type"`
	Content     pgtype.JSONB  `db:"content" json:"content"`
}

func (q *Queries) GetPreviousTopics(ctx context.Context, arg *GetPreviousTopicsParams) ([]*GetPreviousTopicsRow, error) {
	rows, err := q.db.Query(ctx, getPreviousTopics, arg.UserID, arg.Type)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*GetPreviousTopicsRow
	for rows.Next() {
		var i GetPreviousTopicsRow
		if err := rows.Scan(
			&i.ID,
			&i.NextTopicID,
			&i.Type,
			&i.Content,
		); err != nil {
			return nil, err
		}
		items = append(items, &i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTopic = `-- name: GetTopic :one
SELECT id, next_topic_id, type, content
FROM topics
WHERE id = $1
`

type GetTopicRow struct {
	ID          int64         `db:"id" json:"id"`
	NextTopicID sql.NullInt64 `db:"next_topic_id" json:"next_topic_id"`
	Type        string        `db:"type" json:"type"`
	Content     pgtype.JSONB  `db:"content" json:"content"`
}

func (q *Queries) GetTopic(ctx context.Context, id int64) (*GetTopicRow, error) {
	row := q.db.QueryRow(ctx, getTopic, id)
	var i GetTopicRow
	err := row.Scan(
		&i.ID,
		&i.NextTopicID,
		&i.Type,
		&i.Content,
	)
	return &i, err
}

const updateTopicRelation = `-- name: UpdateTopicRelation :exec
UPDATE topics
SET next_topic_id = $2
WHERE id = $1
`

type UpdateTopicRelationParams struct {
	ID          int64         `db:"id" json:"id"`
	NextTopicID sql.NullInt64 `db:"next_topic_id" json:"next_topic_id"`
}

func (q *Queries) UpdateTopicRelation(ctx context.Context, arg *UpdateTopicRelationParams) error {
	_, err := q.db.Exec(ctx, updateTopicRelation, arg.ID, arg.NextTopicID)
	return err
}
