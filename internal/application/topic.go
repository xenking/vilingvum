package application

import (
	"context"
	"math/rand"
	"strings"
	"time"

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

		topic := b.topics.Get(currentTopicID)

		return c.Send(b.prepareTopic(ctx, topic, user.ID))
	}
}

func loadTopics(ctx context.Context, db *database.DB) (*domain.Topics, error) {
	dbTopic, err := db.GetTopics(ctx)
	if err != nil {
		return nil, err
	}

	tt := make([]*domain.Topic, len(dbTopic))
	for i, topic := range dbTopic {
		tt[i] = &domain.Topic{}
		err = json.Unmarshal(topic.Content.Bytes, tt[i])
		if err != nil {
			return nil, err
		}
		tt[i].ID = topic.ID
		tt[i].NextTopicID = topic.NextTopicID.Int64
		tt[i].Type = domain.TopicType(topic.Type)
	}

	return &domain.Topics{
		Data: tt,
	}, nil
}

func filterVideoTopics(tt *domain.Topics) *domain.Topics {
	topics := &domain.Topics{
		Index: make([]int, len(tt.Data)),
	}

	for i, topic := range tt.Data {
		topics.Index[i] = len(topics.Data)
		if topic.Type == domain.TopicTypeVideo {
			topics.Data = append(topics.Data, topic)
		}
	}

	return topics
}

func (b *Bot) prepareTopic(ctx context.Context, topic *domain.Topic, userID int64) (*domain.Topic, *tele.ReplyMarkup) {
	topic.Raw.Reset()

	b.users.SetTopicID(userID, topic.ID)

	//nolint:exhaustive
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

		answers = append(answers, nextBtn) //nolint:makezero
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
		case (num == 2 && sumLen > 37) || (num == 3 && sumLen > 40) || num >= 4:
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

		for _, buttons := range c.Callback().Message.ReplyMarkup.InlineKeyboard {
			for _, btn := range buttons {
				if strings.Contains(btn.Data, "answered") {
					return nil
				}
			}
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
				retryTopics := rt.(map[int64]*domain.Topic) //nolint:forcetypeassert
				if _, exists := retryTopics[topic.ID]; !exists {
					retryTopics[topic.ID] = topic
				}
			}
		}

		for i, buttons := range menu.InlineKeyboard {
			for j, btn := range buttons {
				b.Handle(&buttons[j], emptyCallback)

				if btn.Text != answer.Text {
					continue
				}

				btn.Data = "answered"
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

		nextTopic := b.topics.Get(topic.NextTopicID)

		//nolint:exhaustive
		switch nextTopic.Type {
		case domain.TopicTypeVideo, domain.TopicTypeTestReport:
			rt, exists := b.retryTopics.Get(user.ID)
			retry := false

			retryTopics, ok := rt.(map[int64]*domain.Topic)
			if exists && ok && len(retryTopics) > 0 {
				for key, retryTopic := range retryTopics {
					nextTopic = retryTopic
					retryTopic.NextTopicID = topic.NextTopicID
					delete(retryTopics, key)
					retry = true

					break
				}
			}

			if !retry {
				err = c.Send(getRandomCompliment())
				if err != nil {
					return err
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

		return c.Send(b.prepareTopic(ctx, nextTopic, user.ID))
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

		nextTopic := b.topics.Get(topic.NextTopicID)

		return c.Send(b.prepareTopic(ctx, nextTopic, user.ID))
	}
}

func (b *Bot) HandleGetVideoTopics(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.ResetAction(c)

		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		lastTopicID, err := b.db.GetLatestTopicID(ctx, user.ID)
		if err != nil {
			return err
		}

		opts := &tele.SendOptions{
			ParseMode: tele.ModeMarkdownV2,
			Protected: true,
		}

		var sb strings.Builder

		if lastTopicID < 2 {
			sb.WriteString("You have no previous topics")
		} else {
			idx := b.videoTopics.Index[lastTopicID]

			sb.WriteString("Previous topics:\n")

			topicsRaw := b.paginationVideoTopics(1, int64(idx))
			sb.WriteString(topicsRaw)

			if idx/PaginationPageSize > 0 {
				opts.ReplyMarkup = b.paginationMenu(ctx, 1, int64(idx), b.HandleCallbackVideoTopics)
			}
		}

		return c.Send(domain.EscapeString(sb.String()), opts)
	}
}

func (b *Bot) paginationVideoTopics(page, lastIdx int64) string {
	var sb strings.Builder

	page--

	end := (page + 1) * PaginationPageSize
	if lastIdx < end {
		end = lastIdx
	}

	for i := page * PaginationPageSize; i < end && i < int64(len(b.videoTopics.Data)); i++ {
		sb.WriteString(utils.WriteUint(i + 1))
		sb.WriteString(". ")
		sb.WriteString(b.videoTopics.Data[i].Text)
		sb.WriteByte('\n')
		sb.WriteString(b.videoTopics.Data[i].VideoURL)

		if i != end-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

func (b *Bot) HandleCallbackVideoTopics(ctx context.Context, page, lastIdx int64) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		opts := &tele.SendOptions{
			ParseMode: tele.ModeMarkdownV2,
			Protected: true,
		}

		var sb strings.Builder
		sb.WriteString("Previous topics:\n")

		topicsRaw := b.paginationVideoTopics(page, lastIdx)
		sb.WriteString(topicsRaw)

		opts.ReplyMarkup = b.paginationMenu(ctx, page, lastIdx, b.HandleCallbackVideoTopics)

		return c.Edit(domain.EscapeString(sb.String()), opts)
	}
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec

func getRandomCompliment() string {
	compliments := []string{
		"Excellent! 🎉", "It’s a way to go! 🎉", "Good job! 🎉", "Second to none! 🎉",
		"Fantastic results! 🎉", "I’m proud of you! 🎉",
	}
	idx := rnd.Intn(len(compliments))

	return compliments[idx]
}
