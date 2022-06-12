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

		reply := menu.Guest
		user := b.users.Get(c.Sender().ID)

		if user == nil {
			var err error
			user, err = b.users.Add(ctx, c.Sender())
			if err != nil {
				return err
			}

			return c.Send("Hello there! I'm a bot", reply)
		}

		if b.validSubscription(user) {
			reply = menu.Main
		}

		// TODO: invite system?
		//if c.Message().Payload != "" {
		//	active := time.Now().AddDate(0, 1, 0)
		//	err := b.users.UpdateLicense(user.ID, active)
		//	if err != nil {
		//		return err
		//	}
		//
		//	return c.Send("Your license has been activated", menu.Main)
		//}

		return c.Send(fmt.Sprintf("Hey %s!", user.Name), reply)
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

func (b *Bot) validSubscription(user *domain.User) bool {
	if user == nil {
		return false
	}

	if user.ActiveUntil == nil {
		return false
	}

	return user.ActiveUntil.After(time.Now())
}
