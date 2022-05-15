package logger

import (
	"os"

	"github.com/phuslu/log"

	"tgbot/config"
)

var Log *log.Logger

func SetGlobal(logger *log.Logger) {
	Log = logger
}

func New(cfg *config.LoggerConfig) *log.Logger {
	level := log.ParseLevel(cfg.Level)
	log.DefaultLogger.SetLevel(level)

	w := &log.ConsoleWriter{
		ColorOutput:    true,
		QuoteString:    true,
		EndWithMessage: true,
		Writer:         os.Stdout,
	}

	return &log.Logger{
		Level:  level,
		Caller: cfg.WithCaller,
		Writer: w,
	}
}

func NewModule(name string) *log.Logger {
	ctx := log.NewContext(nil).Str("module", name).Value()

	return &log.Logger{
		Level:          Log.Level,
		Caller:         Log.Caller,
		FullpathCaller: Log.FullpathCaller,
		TimeField:      Log.TimeField,
		TimeFormat:     Log.TimeFormat,
		Context:        ctx,
		Writer:         Log.Writer,
	}
}
