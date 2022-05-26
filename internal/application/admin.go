package application

import (
	"context"

	tele "gopkg.in/telebot.v3"
)

// TODO: implement this
func (b *Bot) GetUserInfo(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		return c.Send("Not implemented")
	}
}
