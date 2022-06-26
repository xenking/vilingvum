package domain

import (
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/pkg/utils"
)

type FeedbackMsg struct {
	UserID int64
}

func (msg *FeedbackMsg) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	var sb strings.Builder

	sb.WriteString("Feedback report\nFrom user: *")
	sb.WriteString(utils.WriteUint(msg.UserID))
	sb.WriteString("*")

	options.ParseMode = tele.ModeMarkdownV2

	return bot.Send(recipient, escape.Replace(sb.String()), options)
}
