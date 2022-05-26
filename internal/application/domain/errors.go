package domain

import tele "gopkg.in/telebot.v3"

func NewError(err error) *Error {
	return &Error{
		Err: err,
	}
}

type Error struct {
	Err     error
	Message string
}

func (e Error) Error() string {
	return e.Err.Error()
}

func (e Error) Send(bot *tele.Bot, recipient tele.Recipient, options *tele.SendOptions) (*tele.Message, error) {
	if e.Message == "" {
		e.Message = "Unexpected error occurred. Contact the bot owner.\nError: " + e.Err.Error()
	}

	return bot.Send(recipient, e.Message, options)
}
