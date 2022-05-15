package middlewares

import (
	"github.com/cornelk/hashmap"
	"github.com/goccy/go-json"
	"github.com/phuslu/log"
	tele "gopkg.in/telebot.v3"
)

func IsUser(users *hashmap.HashMap) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if _, exists := users.Get(c.Sender().ID); !exists {
				return nil
			}

			return next(c)
		}
	}
}

func Logger(l *log.Logger) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			data, _ := json.Marshal(c.Update())
			l.Debug().RawJSON("data", data).Msg("update")
			return next(c)
		}
	}
}
