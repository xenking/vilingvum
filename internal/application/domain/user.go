package domain

import "time"

type User struct {
	Name        string
	Settings    UserSetting
	ActiveUntil *time.Time
	ID          int64
	IsAdmin     bool
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type UserSetting struct{}

type UserAnswer struct {
	TopicType TopicType   `json:"topic_type"`
	Text      string      `json:"text"`
	Answer    interface{} `json:"answer,omitempty"`
}
