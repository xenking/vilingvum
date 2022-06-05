package application

import (
	"context"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/internal/application/domain"
)

func (b *Bot) GetDict(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.ResetAction(c)

		user := b.users.Get(c.Sender().ID)
		if user == nil {
			return c.Send("You are not registered")
		}

		topicID := b.users.GetTopicID(user.ID)

		dd, dbErr := b.db.GetDictionary(ctx, topicID)
		if dbErr != nil {
			return dbErr
		}

		dict := make(domain.Dictionary, len(dd))
		for i, r := range dd {
			dict[i] = domain.DictRecord{
				Word:    r.Word,
				Meaning: r.Meaning,
			}
		}

		return c.Send(dict)
	}
}
