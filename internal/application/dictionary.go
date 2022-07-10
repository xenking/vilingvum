package application

import (
	"context"
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/internal/application/domain"
	"github.com/xenking/vilingvum/pkg/utils"
)

func loadDictionary(ctx context.Context, lastTopicID int64, db *database.DB) (*domain.Dictionary, error) {
	dbDict, err := db.GetDictionary(ctx)
	if err != nil {
		return nil, err
	}

	dict := &domain.Dictionary{
		Data:  make([]domain.DictRecord, len(dbDict)),
		Index: make([]int, lastTopicID),
	}

	prevTopicID := dbDict[0].TopicID
	for i, r := range dbDict {
		dict.Data[i] = domain.DictRecord{
			Word:    r.Word,
			Meaning: r.Meaning,
			TopicID: r.TopicID,
			ID:      r.ID,
		}

		for id := prevTopicID; id < r.TopicID; id++ {
			dict.Index[id] = i
			prevTopicID = r.TopicID
		}
	}

	for id := prevTopicID; id < lastTopicID; id++ {
		dict.Index[id] = len(dict.Data)
	}

	return dict, nil
}

func (b *Bot) GetDict(ctx context.Context) tele.HandlerFunc {
	return func(c tele.Context) error {
		b.ResetAction(c)

		user := b.users.Get(c.Sender().ID)
		if user == nil {
			return c.Send("You are not registered")
		}

		topicID := b.users.GetTopicID(user.ID)

		opts := &tele.SendOptions{
			ParseMode: tele.ModeMarkdownV2,
			Protected: true,
		}

		var sb strings.Builder

		idx := b.dict.Index[topicID] - 1

		if idx < 1 {
			sb.WriteString("Your dictionary is empty")
		} else {
			sb.WriteString("Your dictionary:\n")

			dictRaw := b.paginationDict(1, int64(idx))
			sb.WriteString(dictRaw)

			if idx/PaginationPageSize > 0 {
				opts.ReplyMarkup = b.paginationMenu(ctx, 1, int64(idx), b.HandleCallbackDictPagination)
			}
		}

		return c.Send(domain.EscapeString(sb.String()), opts)
	}
}

func (b *Bot) paginationDict(page, lastIdx int64) string {
	var sb strings.Builder

	page--

	end := (page + 1) * PaginationPageSize
	if lastIdx < end {
		end = lastIdx
	}

	for i := page * PaginationPageSize; i < end && i < int64(len(b.dict.Data)); i++ {
		sb.WriteString(utils.WriteUint(b.dict.Data[i].ID))
		sb.WriteString(". *")
		sb.WriteString(b.dict.Data[i].Word)
		sb.WriteString("*: ")
		sb.WriteString(b.dict.Data[i].Meaning)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (b *Bot) HandleCallbackDictPagination(ctx context.Context, page, lastIdx int64) tele.HandlerFunc {
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
		sb.WriteString("Your dictionary:\n")

		dictRaw := b.paginationDict(page, lastIdx)
		sb.WriteString(dictRaw)

		opts.ReplyMarkup = b.paginationMenu(ctx, page, lastIdx, b.HandleCallbackDictPagination)

		return c.Edit(domain.EscapeString(sb.String()), opts)
	}
}
