package application

import (
	"context"
	"fmt"
	"time"

	"github.com/phuslu/log"
	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/internal/application/menu"
)

func (b *Bot) HandleStart(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.ResetAction(c)

		user := b.users.Get(c.Sender().ID)

		payload := c.Message().Payload
		if user != nil && payload == "" {
			return c.Send(fmt.Sprintf("Hey %s!", user.Name), menu.Main)
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
				return err
			}

			return c.Send("Your license has been activated", menu.Main)
		}

		return c.Send("Hello there! I'm a bot", menu.Main)
	}
}

func (b *Bot) HandleAbout(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.ResetAction(c)

		return c.Send("I'm a bot that can help you to learn English words")
	}
}

func (b *Bot) HandleFeedback(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		t := &domain.Topic{
			Text: "Write your feedback here and I will send it to the teacher",
			Type: domain.TopicTypeTestReport,
		}
		b.actions.Set(c.Sender().ID, domain.ActionFeedback)

		return c.Send(b.prepareTopic(ctx, t))
	}
}

func (b *Bot) HandleError(ctx context.Context) func(error, tele.Context) {
	return func(err error, c tele.Context) {
		log.Error().Err(err).Msg("global")

		sErr := c.Send(domain.NewError(err))
		panic(sErr)
	}
}

func (b *Bot) getUser(c tele.Context) *domain.User {
	return b.users.Get(c.Sender().ID)
}
