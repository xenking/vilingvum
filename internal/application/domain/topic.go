package domain

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

type TopicType string

const (
	TopicVideo       TopicType = "video"
	TopicRepeatVideo TopicType = "repeat"
	TopicTestTitle   TopicType = "test_title"
	TopicQuestion    TopicType = "question"
	TopicTest        TopicType = "test"
	TopicTestReport  TopicType = "report"
)

type Topic struct {
	Text       string        `json:"text,omitempty"`
	VideoURL   []string      `json:"video_url,omitempty"`
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

func (t *Topic) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	options.ParseMode = tele.ModeMarkdownV2
	options.Protected = true

	return bot.Send(recipient, escape.Replace(t.Raw.String()), options)
}

type Topics struct {
	Data  []*Topic
	Index []int
}

func (tt *Topics) Get(id int64) *Topic {
	return tt.Data[id-1]
}
