package application

import (
	"context"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/internal/application/menu"
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
		case domain.ActionFeedback:
			if media := c.Message().Media(); media != nil {
				return c.Send("Media not supported yet, use text instead")
			}
			b.actions.Del(user.ID)

			for _, adminID := range b.forwardIDs {
				_, err := b.Send(adminID, &domain.FeedbackMsg{
					UserID: user.ID,
				})
				if err != nil {
					return err
				}

				err = c.ForwardTo(adminID)
				if err != nil {
					return err
				}
			}

			return c.Send("Thank you for your feedback!", menu.Main)
		case domain.ActionTestReport:
			media := c.Message().Media()
			if media == nil {
				return c.Send("Please send a video or audio")
			}

			switch media.MediaType() {
			case "video", "audio", "voice":
				b.actions.Del(user.ID)

				for _, adminID := range b.forwardIDs {
					if err := c.ForwardTo(adminID); err != nil {
						return err
					}
				}
			}

			b.users.NextTopicID(user.ID)

			return c.Send("Thank you. I will resend your message to the teacher", menu.Main)
		}

		return nil
	}
}

func (b *Bot) ResetAction(c tele.Context) {
	b.actions.Del(c.Sender().ID)
}
