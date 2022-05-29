package domain

import (
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/pkg/utils"
)

type TopicType string

const (
	TopicTypeVideo    TopicType = "video"
	TopicTypeQuestion TopicType = "question"
	TopicTypeTest     TopicType = "test"
)

type Topic struct {
	Text       string        `json:"text,omitempty"`
	VideoURL   string        `json:"video_url,omitempty"`
	Question   string        `json:"question,omitempty"`
	NextButton string        `json:"next_button,omitempty"`
	Answers    []TopicAnswer `json:"answers,omitempty"`

	ID          int64           `json:"-"`
	NextTopicID int64           `json:"-"`
	Type        TopicType       `json:"-"`
	Raw         strings.Builder `json:"-"`
}

type TopicAnswer struct {
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}

var escape = strings.NewReplacer(".", `\.`, "#", `\#`, "=", `\=`, "+", `\+`, `-`, `\-`, `_`, `\_`)

func (t *Topic) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	options.ParseMode = tele.ModeMarkdownV2

	return bot.Send(recipient, escape.Replace(t.Raw.String()), options)
}

type Topics []*Topic

func (tt Topics) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	options.ParseMode = tele.ModeMarkdownV2

	var sb strings.Builder

	for i, topic := range tt {
		sb.WriteString(utils.WriteUint(int64(i + 1)))
		sb.WriteString(". ")
		sb.WriteString(topic.Raw.String())
		if i != len(tt)-1 {
			sb.WriteString("\n")
		}
	}

	return bot.Send(recipient, escape.Replace(sb.String()), options)
}
