package domain

import (
	"time"

	"github.com/xenking/vilingvum/pkg/utils"
)

type User struct {
	ActiveUntil *time.Time
	Name        string
	ID          int64
	IsAdmin     bool
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type UserAnswer struct {
	TopicType TopicType   `json:"topic_type"`
	Text      string      `json:"text"`
	Answer    interface{} `json:"answer,omitempty"`
}

type ForwardID int64

func (id ForwardID) Recipient() string {
	return utils.WriteUint(int64(id))
}
