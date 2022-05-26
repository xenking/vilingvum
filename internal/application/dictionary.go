package application

import (
	"context"

	tele "gopkg.in/telebot.v3"
)

func (b *Bot) GetDict(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.users.Get(c.Sender().ID)
		if user == nil {
			return c.Send("You are not registered")
		}

		_ = c.Send("Dictionary: ")

		return c.Send(user.Dictionary)
	}
}
