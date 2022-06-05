package domain

import (
	"strings"

	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/pkg/utils"
)

type Dictionary []DictRecord

type DictRecord struct {
	Word    string
	Meaning string
}

func (dict Dictionary) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	var sb strings.Builder

	if len(dict) > 0 {
		sb.WriteString("Your dictionary:\n")
	} else {
		sb.WriteString("Your dictionary is empty")
	}

	for i, r := range dict {
		sb.WriteString(utils.WriteUint(int64(i + 1)))
		sb.WriteString(". *")
		sb.WriteString(r.Word)
		sb.WriteString("*: ")
		sb.WriteString(r.Meaning)
		sb.WriteString("\n")
	}

	options.ParseMode = tele.ModeMarkdownV2
	options.Protected = true

	return bot.Send(recipient, escape.Replace(sb.String()), options)
}
