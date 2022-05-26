package application

import (
	"context"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/internal/application/domain"
)

func (b *Bot) OnAction(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.getUser(c)
		if user == nil {
			return nil
		}

		action, ok := b.actions.Get(user.ID)
		if !ok {
			return nil
		}

		switch action {
		case domain.ActionTestReport:
			return c.Send("Thank you. I will send your report to the teacher")
		}

		return nil
	}
}
