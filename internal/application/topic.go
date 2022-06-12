package application

import (
	"context"
	"math/rand"

	"github.com/goccy/go-json"
	"github.com/jackc/pgtype"
	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/pkg/utils"
)

func (b *Bot) HandleGetCurrentTopic(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.ResetAction(c)

		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		currentTopicID := b.users.GetTopicID(user.ID)
		if currentTopicID < 0 {
			return c.Send("No current topic")
		}

		if !b.validSubscription(user) && currentTopicID >= domain.DemoTopicID-1 {
			return c.Send("All topics are available for subscription")
		}

		topic, err := b.loadTopic(ctx, currentTopicID)
		if err != nil {
			return err
		}

		return c.Send(b.prepareTopic(ctx, topic))
	}
}

func (b *Bot) loadTopic(ctx context.Context, topicID int64) (*domain.Topic, error) {
	dbTopic, err := b.db.GetTopic(ctx, topicID)
	if err != nil {
		return nil, err
	}

	t := &domain.Topic{
		ID:          dbTopic.ID,
		NextTopicID: dbTopic.NextTopicID.Int64,
		Type:        domain.TopicType(dbTopic.Type),
	}
	err = json.Unmarshal(dbTopic.Content.Bytes, t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (b *Bot) prepareTopic(ctx context.Context, topic *domain.Topic) (*domain.Topic, *tele.ReplyMarkup) {
	topic.Raw.Reset()

	switch topic.Type {
	case domain.TopicTypeVideo:
		topic.Raw.WriteString("Next topic *")
		topic.Raw.WriteString(topic.Text)
		topic.Raw.WriteString("*:\n")
		topic.Raw.WriteString(topic.VideoURL)
	case domain.TopicTypeQuestion, domain.TopicTypeTest:
		topic.Raw.WriteString(topic.Text)
		topic.Raw.WriteString(":\n*")
		topic.Raw.WriteString(topic.Question)
		topic.Raw.WriteByte('*')
	default:
		topic.Raw.WriteString(topic.Text)
	}

	replyAnswers := &tele.ReplyMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
		ForceReply:      true,
	}

	answers := make([]tele.Btn, len(topic.Answers))
	for i, a := range topic.Answers {
		btn := replyAnswers.Data(a.Text, "answer_"+utils.WriteUint(int64(i)))

		b.Handle(&btn, b.HandleCallbackAnswer(ctx, topic, a, replyAnswers))

		answers[i] = btn
	}

	rand.Shuffle(len(answers), func(i, j int) {
		answers[i], answers[j] = answers[j], answers[i]
	})

	if topic.NextButton != "" {
		nextBtn := replyAnswers.Data(topic.NextButton, "next")

		b.Handle(&nextBtn, b.HandleCallbackNextTopic(ctx, topic))

		answers = append(answers, nextBtn)
	}

	var rows []tele.Row

	var sumLen, lastSplit, num int
	for i := range answers {
		num++
		sumLen += len(answers[i].Text)
		switch {
		case len(answers[i].Text) >= 30:
			if lastSplit < i {
				rows = append(rows, answers[lastSplit:i])
			}
			rows = append(rows, tele.Row{answers[i]})
			sumLen = 0
			lastSplit = i + 1
			num = 0
		case (num == 2 && sumLen > 37) || (num == 3 && sumLen > 47) || num >= 4:
			rows = append(rows, answers[lastSplit:i])
			sumLen = len(answers[i].Text)
			lastSplit = i
			num = 1
		}
	}

	if lastSplit < len(answers) {
		rows = append(rows, answers[lastSplit:])
	}

	replyAnswers.Inline(rows...)

	return topic, replyAnswers
}

func (b *Bot) HandleCallbackAnswer(ctx context.Context, topic *domain.Topic, answer domain.TopicAnswer, menu *tele.ReplyMarkup) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		buf, err := json.Marshal(domain.UserAnswer{
			TopicType: topic.Type,
			Text:      topic.Text,
			Answer:    answer,
		})
		if err != nil {
			return err
		}

		err = b.db.InsertUserAnswer(ctx, &database.InsertUserAnswerParams{
			UserID:  user.ID,
			TopicID: topic.ID,
			Response: pgtype.JSONB{
				Bytes:  buf,
				Status: pgtype.Present,
			},
		})
		if err != nil {
			return err
		}

		if !answer.Correct {
			rt, loaded := b.retryTopics.GetOrInsert(user.ID, map[int64]*domain.Topic{topic.ID: topic})
			if loaded {
				retryTopics := rt.(map[int64]*domain.Topic)
				if _, exists := retryTopics[topic.ID]; !exists {
					retryTopics[topic.ID] = topic
				}
			}
		}

		for i, buttons := range menu.InlineKeyboard {
			for j, btn := range buttons {
				if btn.Text != answer.Text {
					continue
				}

				if !answer.Correct {
					btn.Text = "❌ " + btn.Text
					err = c.Respond(&tele.CallbackResponse{
						Text: "You answered incorrectly",
					})
				} else {
					btn.Text = "✅ " + btn.Text
					err = c.Respond(&tele.CallbackResponse{
						Text: "You answered correctly",
					})
				}
				if err != nil {
					return err
				}

				menu.InlineKeyboard[i][j] = btn
			}
		}

		err = c.Edit(menu)
		if err != nil {
			return err
		}

		nextTopic, err := b.loadTopic(ctx, topic.NextTopicID)
		if err != nil {
			return err
		}

		b.users.SetTopicID(user.ID, nextTopic.ID)

		switch nextTopic.Type {
		case domain.TopicTypeVideo, domain.TopicTypeTestReport:
			rt, exists := b.retryTopics.Get(user.ID)

			retryTopics, ok := rt.(map[int64]*domain.Topic)
			if exists && ok && len(retryTopics) > 0 {
				for key, retryTopic := range retryTopics {
					nextTopic = retryTopic
					retryTopic.NextTopicID = topic.NextTopicID
					delete(retryTopics, key)

					break
				}
			}
		}

		if nextTopic.Type == domain.TopicTypeTestReport {
			b.actions.Set(user.ID, domain.ActionTestReport)
		}

		c.DeleteAfter(domain.TopicDeleteDelay)

		if !b.validSubscription(user) && topic.NextTopicID >= domain.DemoTopicID {
			return c.Send("All topics are available for subscription")
		}

		return c.Send(b.prepareTopic(ctx, nextTopic))
	}
}

func (b *Bot) HandleCallbackNextTopic(ctx context.Context, topic *domain.Topic) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		buf, err := json.Marshal(domain.UserAnswer{
			TopicType: topic.Type,
			Text:      topic.Text,
		})
		if err != nil {
			return err
		}

		err = b.db.InsertUserAnswer(ctx, &database.InsertUserAnswerParams{
			UserID:  user.ID,
			TopicID: topic.ID,
			Response: pgtype.JSONB{
				Bytes:  buf,
				Status: pgtype.Present,
			},
		})
		if err != nil {
			return err
		}

		if !b.validSubscription(user) && topic.NextTopicID >= domain.DemoTopicID {
			return c.Send("All topics are available for subscription")
		}

		nextTopic, err := b.loadTopic(ctx, topic.NextTopicID)
		if err != nil {
			return err
		}

		b.users.SetTopicID(user.ID, nextTopic.ID)

		return c.Send(b.prepareTopic(ctx, nextTopic))
	}
}

func (b *Bot) HandleGetPrevTopics(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.ResetAction(c)

		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		tt, err := b.db.GetPreviousTopics(ctx, &database.GetPreviousTopicsParams{
			UserID: user.ID,
			Type:   string(domain.TopicTypeVideo),
		})
		if err != nil {
			return err
		}

		if len(tt) == 0 {
			return c.Send("You have no previous topics")
		}

		topics := make(domain.Topics, len(tt))

		for i, dbTopic := range tt {
			topic := &domain.Topic{
				ID:          dbTopic.ID,
				NextTopicID: dbTopic.NextTopicID.Int64,
				Type:        domain.TopicType(dbTopic.Type),
			}
			err = json.Unmarshal(dbTopic.Content.Bytes, topic)
			if err != nil {
				return err
			}

			topic.Raw.Reset()
			topic.Raw.WriteString(topic.Text)
			topic.Raw.WriteByte('\n')
			topic.Raw.WriteString(topic.VideoURL)

			topics[i] = topic
		}

		return c.Send(topics)
	}
}
