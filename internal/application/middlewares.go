package application

import (
	"github.com/goccy/go-json"
	"github.com/phuslu/log"
	tele "gopkg.in/telebot.v3"

	"github.com/xenking/vilingvum/internal/application/users"
)

func IsUserMiddleware(store *users.Store) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if u := store.Get(c.Sender().ID); u == nil {
				return nil
			}

			return next(c)
		}
	}
}

func LoggerMiddleware(l *log.Logger) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			data, _ := json.Marshal(c.Update())
			l.Debug().RawJSON("data", data).Msg("update")
			return next(c)
		}
	}
}
