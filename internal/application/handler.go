package application

import (
	"context"
	"fmt"
	"time"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/internal/application/menu"
)

func (b *Bot) getUser(c tele.Context) *domain.User {
	return b.users.Get(c.Sender().ID)
}

func (b *Bot) HandleStart(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.users.Get(c.Sender().ID)

		payload := c.Message().Payload
		if user != nil && payload == "" {
			return c.Send(fmt.Sprintf("Hello %s !", user.Name), menu.Main)
		} else if user == nil {
			var err error
			user, err = b.users.Add(ctx, c.Sender())
			if err != nil {
				return err
			}
		}

		if payload != "" {
			active := time.Now().AddDate(0, 1, 0)
			err := b.users.UpdateLicense(user.ID, active)
			if err != nil {
				return c.Send(domain.NewError(err))
			}

			return c.Send("Your license has been activated", menu.Main)
		}

		return c.Send("Hello there! I'm a bot", menu.Main)
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
