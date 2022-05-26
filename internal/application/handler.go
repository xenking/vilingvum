package application

import (
	"context"
	"fmt"
	"github.com/xenking/vilingvum/internal/application/menu"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/internal/application/domain"
)

func (b *Bot) getUser(c tele.Context) *domain.User {
	return b.users.Get(c.Sender().ID)
}

func (b *Bot) HandleStart(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.users.Get(c.Sender().ID)
		if c.Message().Payload != "" && user == nil {
			var err error
			active := time.Now().AddDate(0, 1, 0)

			user, err = b.users.Add(ctx, c.Sender(), active)
			if err != nil {
				return err
			}
		}

		if user != nil {
			return c.Send(fmt.Sprintf("Hello %s !", user.Name), menu.Main)
		}

		return c.Send("Hello there! I'm a bot")
	}
}

func (b *Bot) HandleAbout(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		return c.Send("I'm a bot")
	}
}

func (b *Bot) HandleFeedback(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		return c.Send("Feedback is welcome")
	}
}
