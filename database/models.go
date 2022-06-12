// Code generated by sqlc. DO NOT EDIT.

package database

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

type Dictionary struct {
	ID      int64  `db:"id" json:"id"`
	TopicID int64  `db:"topic_id" json:"topic_id"`
	Word    string `db:"word" json:"word"`
	Meaning string `db:"meaning" json:"meaning"`
}

type Invoice struct {
	Uuid      uuid.UUID    `db:"uuid" json:"uuid"`
	UserID    int64        `db:"user_id" json:"user_id"`
	Payload   pgtype.JSONB `db:"payload" json:"payload"`
	Email     *string      `db:"email" json:"email"`
	ChargeID  *string      `db:"charge_id" json:"charge_id"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
}

type Topic struct {
	ID          int64         `db:"id" json:"id"`
	NextTopicID sql.NullInt64 `db:"next_topic_id" json:"next_topic_id"`
	Type        string        `db:"type" json:"type"`
	Content     pgtype.JSONB  `db:"content" json:"content"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
}

type User struct {
	ID          int64        `db:"id" json:"id"`
	Name        string       `db:"name" json:"name"`
	Username    string       `db:"username" json:"username"`
	Email       *string      `db:"email" json:"email"`
	PhoneNumber *string      `db:"phone_number" json:"phone_number"`
	State       string       `db:"state" json:"state"`
	IsAdmin     bool         `db:"is_admin" json:"is_admin"`
	Settings    pgtype.JSONB `db:"settings" json:"settings"`
	ActiveUntil *time.Time   `db:"active_until" json:"active_until"`
	CreatedAt   time.Time    `db:"created_at" json:"created_at"`
}

type UserAnswer struct {
	ID        int64        `db:"id" json:"id"`
	UserID    int64        `db:"user_id" json:"user_id"`
	TopicID   int64        `db:"topic_id" json:"topic_id"`
	Response  pgtype.JSONB `db:"response" json:"response"`
	UpdatedAt time.Time    `db:"updated_at" json:"updated_at"`
	CreatedAt time.Time    `db:"created_at" json:"created_at"`
}
