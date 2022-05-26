package domain

import (
	tele "gopkg.in/telebot.v3"
	"strings"
)

type User struct {
	Name       string
	Settings   UserSetting
	Dictionary UserDictionary
	ID         int64
	IsAdmin    bool
}

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
)

type UserSetting struct{}

type UserDictionary map[string]string

func (u UserDictionary) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	var sb strings.Builder
	for k, v := range u {
		sb.WriteString("1. *")
		sb.WriteString(k)
		sb.WriteString("*: ")
		sb.WriteString(v)
	}

	options.ParseMode = tele.ModeMarkdownV2

	return bot.Send(recipient, sb.String(), options)
}
