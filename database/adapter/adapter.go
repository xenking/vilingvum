package adapter

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/phuslu/log"
)

type Logger struct {
	logger *log.Logger
}

// NewPgLogger accepts a log.Logger as input and returns a new custom pgx
// logging fascade as output.
func NewPgLogger(logger *log.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Log(_ context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	var ll log.Level
	switch level {
	case pgx.LogLevelNone:
		ll = log.TraceLevel
	case pgx.LogLevelError:
		ll = log.ErrorLevel
	case pgx.LogLevelWarn:
		ll = log.WarnLevel
	case pgx.LogLevelInfo:
		ll = log.InfoLevel
	case pgx.LogLevelDebug:
		ll = log.DebugLevel
	default:
		ll = log.DebugLevel
	}

	l.logger.WithLevel(ll).Fields(data).Msg(msg)
}
