package domain

import (
	"time"

	"github.com/xenking/vilingvum/pkg/utils"
)

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

type ForwardID int64

func (id ForwardID) Recipient() string {
	return utils.WriteUint(int64(id))
}
