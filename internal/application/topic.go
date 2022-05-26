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
	topic.Raw.WriteString(topic.Text)
	topic.Raw.WriteByte('\n')
	topic.Raw.WriteByte('*')
	topic.Raw.WriteString(topic.Question)
	topic.Raw.WriteByte('*')

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

	if len(answers) > 5 {
		split := len(answers) / 2
		rows = append(rows, answers[:split], answers[split:])
	} else {
		rows = append(rows, answers)
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

		a, err := json.Marshal(answer)
		if err != nil {
			return c.Send(domain.NewError(err))
		}

		err = b.db.InsertUserAnswer(ctx, &database.InsertUserAnswerParams{
			UserID:  user.ID,
			TopicID: topic.ID,
			Response: pgtype.JSONB{
				Bytes:  a,
				Status: pgtype.Present,
			},
		})
		if err != nil {
			return c.Send(domain.NewError(err))
		}

		if !answer.Correct {
			inc, loaded := b.retryTopics.GetOrInsert(user.ID, []*domain.Topic{topic})
			if loaded {
				incorrect := inc.([]*domain.Topic)
				incorrect = append(incorrect, topic)
				b.retryTopics.Set(user.ID, incorrect)
			}
		}

		nextTopic, err := b.loadTopic(ctx, topic.NextTopicID)
		if err != nil {
			return err
		}

		if nextTopic.Type != domain.TopicTypeQuestion {
			rt, exists := b.retryTopics.Get(user.ID)
			if retryTopics := rt.([]*domain.Topic); exists && len(retryTopics) > 0 {
				nextTopic = retryTopics[0]
				if len(retryTopics) > 1 {
					b.retryTopics.Set(user.ID, retryTopics[1:])
				} else {
					b.retryTopics.Del(user.ID)
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

		nextTopic, err := b.loadTopic(ctx, topic.NextTopicID)
		if err != nil {
			return err
		}

		return c.Send(b.prepareTopic(ctx, nextTopic))
	}
}

func (b *Bot) HandleGetPrevTopics(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		user := b.getUser(c)
		if user == nil {
			return c.Send("You are not registered")
		}

		topics, err := b.db.GetPreviousTopics(ctx, &database.GetPreviousTopicsParams{
			UserID: user.ID,
			Type:   string(domain.TopicTypeVideo),
		})
		if err != nil {
			return c.Send(domain.NewError(err))
		}

		if len(topics) == 0 {
			return c.Send("You have no previous topics")
		}

		return c.Send("Your previous topics:", topics)
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
