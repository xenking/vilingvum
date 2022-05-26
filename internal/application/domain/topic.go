package domain

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

type TopicType string

const (
	TopicTypeVideo    TopicType = "video"
	TopicTypeQuestion TopicType = "question"
	TopicTypeTest     TopicType = "test"
)

type Topic struct {
	ID          int64
	NextTopicID int64
	Type        TopicType
	Text        string        `json:"text"`
	VideoURL    string        `json:"video_url,omitempty"`
	Question    string        `json:"question,omitempty"`
	NextButton  string        `json:"next_button,omitempty"`
	Answers     []TopicAnswer `json:"answers,omitempty"`

	Raw strings.Builder
}

type TopicAnswer struct {
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}

func (t *Topic) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	if t.Type == TopicTypeVideo {
		file := &tele.Video{
			File:    tele.FromURL(t.VideoURL),
			Caption: t.Text,
		}

		return bot.Send(recipient, file, options)
	}

	options.ParseMode = tele.ModeMarkdownV2

	return bot.Send(recipient, t.Raw.String(), options)
}
