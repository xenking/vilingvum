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
		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		currentTopicID := b.users.GetTopicID(user.ID)
		if currentTopicID < 0 {
			return c.Send("No current topic")
		}

		topic, err := b.loadTopic(ctx, currentTopicID)
		if err != nil {
			return c.Send(domain.NewError(err))
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
	for i, answer := range answers {
		num++
		sumLen += len(answer.Text)
		switch {
		case len(answer.Text) >= 30:
			if lastSplit < i {
				rows = append(rows, answers[lastSplit:i])
			}
			rows = append(rows, tele.Row{answer})
			sumLen = 0
			lastSplit = i + 1
			num = 0
		case (num == 2 && sumLen > 37) || (num == 3 && sumLen > 47) || num >= 4:
			rows = append(rows, answers[lastSplit:i])
			sumLen = len(answer.Text)
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
			return c.Send(domain.NewError(err))
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
			return c.Send(domain.NewError(err))
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

		nextTopic, err := b.loadTopic(ctx, topic.NextTopicID)
		if err != nil {
			return err
		}

		b.users.SetTopicID(user.ID, nextTopic.ID)

		if nextTopic.Type != domain.TopicTypeQuestion {
			rt, exists := b.retryTopics.Get(user.ID)
			if retryTopics, ok := rt.(map[int64]*domain.Topic); exists && ok && len(retryTopics) > 0 {
				for key, retryTopic := range retryTopics {
					nextTopic = retryTopic
					retryTopic.NextTopicID = topic.NextTopicID
					delete(retryTopics, key)

					break
				}
			}
		}

		err = c.Send(b.prepareTopic(ctx, nextTopic))
		if err != nil {
			return c.Send(domain.NewError(err))
		}

		return c.Delete()
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
			return c.Send(domain.NewError(err))
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
			return c.Send(domain.NewError(err))
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
		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		tt, err := b.db.GetPreviousTopics(ctx, &database.GetPreviousTopicsParams{
			UserID: user.ID,
			Type:   string(domain.TopicTypeVideo),
		})
		if err != nil {
			return c.Send(domain.NewError(err))
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

//
//func (b *Bot) onReadPost(ctx context.Context, menu *tele.ReplyMarkup) tele.HandlerFunc {
//	return func(c tele.Context) error {
//		fields := strings.Split(c.Data(), "|")
//		if len(fields) == 0 {
//			return nil
//		}
//
//		postID, err := utils.ParseUint(fields[len(fields)-1])
//		if err != nil {
//			return c.Send(Error{Err: err})
//		}
//
//		err = b.db.UpdatePostEntry(ctx, &database.UpdatePostEntryParams{
//			PostID: postID,
//			UserID: c.Sender().ID,
//			Status: PostEntryRead,
//		})
//		if err != nil {
//			return c.Send(Error{Err: err})
//		}
//
//		menu.InlineKeyboard[0][0].Text = "❤️ Read"
//
//		return c.Edit(menu)
//	}
//}
//
//func (b *Bot) onNextPost(ctx context.Context, menu *tele.ReplyMarkup) tele.HandlerFunc {
//	return func(c tele.Context) error {
//		fields := strings.Split(c.Data(), "|")
//		if len(fields) == 0 {
//			return nil
//		}
//
//		postID, err := utils.ParseUint(fields[len(fields)-1])
//		if err != nil {
//			return c.Send(Error{Err: err})
//		}
//
//		user, err := b.getUser(ctx, c.Sender().ID)
//		if err != nil {
//			return c.Send(Error{Err: err})
//		}
//
//		post, err := b.db.GetNextPost(ctx, postID)
//		if err != nil {
//			if errors.Is(err, pgx.ErrNoRows) {
//				user.LastPostID = post.ID
//
//				return c.Send("No more posts")
//			}
//
//			return c.Send(Error{Err: err})
//		}
//
//		user.LastPostID = post.ID
//
//		return b.sendPost(ctx, c, post.ID, post.Content.Bytes)
//	}
//}
