// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: post.sql

package database

import (
	"context"
	"database/sql"

	"github.com/jackc/pgtype"
)

const createPost = `-- name: CreatePost :one
INSERT INTO posts (content)
VALUES ($1)
RETURNING id
`

func (q *Queries) CreatePost(ctx context.Context, content pgtype.JSONB) (int64, error) {
	row := q.db.QueryRow(ctx, createPost, content)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const createPostEntry = `-- name: CreatePostEntry :exec
INSERT INTO post_entries (user_id, post_id, status)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, post_id) DO NOTHING
`

type CreatePostEntryParams struct {
	UserID int64  `db:"user_id" json:"user_id"`
	PostID int64  `db:"post_id" json:"post_id"`
	Status string `db:"status" json:"status"`
}

func (q *Queries) CreatePostEntry(ctx context.Context, arg *CreatePostEntryParams) error {
	_, err := q.db.Exec(ctx, createPostEntry, arg.UserID, arg.PostID, arg.Status)
	return err
}

const getLastPostID = `-- name: GetLastPostID :one
SELECT pe.post_id
FROM post_entries pe
         JOIN users u ON u.id = pe.user_id
WHERE u.id = $1
ORDER BY pe.created_at DESC
LIMIT 1
`

func (q *Queries) GetLastPostID(ctx context.Context, id int64) (int64, error) {
	row := q.db.QueryRow(ctx, getLastPostID, id)
	var post_id int64
	err := row.Scan(&post_id)
	return post_id, err
}

const getNextPost = `-- name: GetNextPost :one
SELECT np.id, np.content
FROM posts p
         JOIN posts np ON np.id = p.next_post_id
WHERE p.id = $1 AND p.id != np.id
`

type GetNextPostRow struct {
	ID      int64        `db:"id" json:"id"`
	Content pgtype.JSONB `db:"content" json:"content"`
}

func (q *Queries) GetNextPost(ctx context.Context, id int64) (*GetNextPostRow, error) {
	row := q.db.QueryRow(ctx, getNextPost, id)
	var i GetNextPostRow
	err := row.Scan(&i.ID, &i.Content)
	return &i, err
}

const getPost = `-- name: GetPost :one
SELECT id, next_post_id, content
FROM posts
WHERE id = $1
`

type GetPostRow struct {
	ID         int64         `db:"id" json:"id"`
	NextPostID sql.NullInt64 `db:"next_post_id" json:"next_post_id"`
	Content    pgtype.JSONB  `db:"content" json:"content"`
}

func (q *Queries) GetPost(ctx context.Context, id int64) (*GetPostRow, error) {
	row := q.db.QueryRow(ctx, getPost, id)
	var i GetPostRow
	err := row.Scan(&i.ID, &i.NextPostID, &i.Content)
	return &i, err
}

const updatePostEntry = `-- name: UpdatePostEntry :exec
UPDATE post_entries
SET status     = $3,
    updated_at = now()
WHERE post_id = $1
  AND user_id = $2
`

type UpdatePostEntryParams struct {
	PostID int64  `db:"post_id" json:"post_id"`
	UserID int64  `db:"user_id" json:"user_id"`
	Status string `db:"status" json:"status"`
}

func (q *Queries) UpdatePostEntry(ctx context.Context, arg *UpdatePostEntryParams) error {
	_, err := q.db.Exec(ctx, updatePostEntry, arg.PostID, arg.UserID, arg.Status)
	return err
}
