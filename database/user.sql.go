// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.13.0
// source: user.sql

package database

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users (id, name, username, invite_code, state, active_until)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, name, username, state, is_admin, settings, invite_code, active_until, created_at
`

type CreateUserParams struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Username    string    `db:"username" json:"username"`
	InviteCode  string    `db:"invite_code" json:"invite_code"`
	State       string    `db:"state" json:"state"`
	ActiveUntil time.Time `db:"active_until" json:"active_until"`
}

func (q *Queries) CreateUser(ctx context.Context, arg *CreateUserParams) (*User, error) {
	row := q.db.QueryRow(ctx, createUser,
		arg.ID,
		arg.Name,
		arg.Username,
		arg.InviteCode,
		arg.State,
		arg.ActiveUntil,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Username,
		&i.State,
		&i.IsAdmin,
		&i.Settings,
		&i.InviteCode,
		&i.ActiveUntil,
		&i.CreatedAt,
	)
	return &i, err
}

const getUser = `-- name: GetUser :one
SELECT id, name, username, state, is_admin, settings, invite_code, active_until, created_at
FROM users
WHERE id = $1
`

func (q *Queries) GetUser(ctx context.Context, id int64) (*User, error) {
	row := q.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Username,
		&i.State,
		&i.IsAdmin,
		&i.Settings,
		&i.InviteCode,
		&i.ActiveUntil,
		&i.CreatedAt,
	)
	return &i, err
}

const isUserExists = `-- name: IsUserExists :one
SELECT true
FROM users
WHERE id = $1
LIMIT 1
`

func (q *Queries) IsUserExists(ctx context.Context, id int64) (bool, error) {
	row := q.db.QueryRow(ctx, isUserExists, id)
	var column_1 bool
	err := row.Scan(&column_1)
	return column_1, err
}

const listActiveUsers = `-- name: ListActiveUsers :many
SELECT id, name, settings, is_admin
FROM users
WHERE active_until > now()
  AND state = 'active'
`

type ListActiveUsersRow struct {
	ID       int64        `db:"id" json:"id"`
	Name     string       `db:"name" json:"name"`
	Settings pgtype.JSONB `db:"settings" json:"settings"`
	IsAdmin  bool         `db:"is_admin" json:"is_admin"`
}

func (q *Queries) ListActiveUsers(ctx context.Context) ([]*ListActiveUsersRow, error) {
	rows, err := q.db.Query(ctx, listActiveUsers)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*ListActiveUsersRow
	for rows.Next() {
		var i ListActiveUsersRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Settings,
			&i.IsAdmin,
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

const listAdmins = `-- name: ListAdmins :many
SELECT id
FROM users
WHERE is_admin = true
`

func (q *Queries) ListAdmins(ctx context.Context) ([]int64, error) {
	rows, err := q.db.Query(ctx, listAdmins)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}