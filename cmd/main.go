package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cristalhq/aconfig"
	"github.com/phuslu/log"
	"github.com/pkg/errors"
)

func main() {
	ctx, cancel := appContext()
	defer cancel()

	if err := runMain(ctx, os.Args[1:]); err != nil {
		log.Error().Err(err).Stack().Msg("main")
	}
}

var (
	errNoCommand      = errors.New("no command provided (serve, migrate, import, help)")
	errNoImportFile   = errors.New("no import file provided")
	errUnimplemented  = errors.New("unimplemented")
	errUnknownCommand = errors.New("unknown command")
)

func runMain(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errNoCommand
	}

	var flags cmdFlags
	if err := loadFlags(&flags, args[1:]); err != nil {
		return err
	}

	cmd := args[0]
	switch cmd {
	case "serve":
		if err := loadFlags(&flags, args[2:]); err != nil {
			return err
		}

		switch subCmd := args[1]; subCmd {
		case "bot":
			return serveBotCmd(ctx, flags)
		case "http":
			return serveHTTPCmd(ctx, flags)
		default:
			return errors.Wrap(errUnknownCommand, cmd+": "+subCmd)
		}
	case "migrate":
		return migrateCmd(ctx, flags)
	case "import":
		if len(args) < 3 || args[2] == "" {
			return errNoImportFile
		}

		return importCmd(ctx, flags)
	case "help":
		panic(errUnimplemented)
	}

	return errors.Wrap(errUnknownCommand, cmd)
}

// appContext returns context that will be canceled on specific OS signals.
func appContext() (context.Context, context.CancelFunc) {
	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}

	ctx, cancel := signal.NotifyContext(context.Background(), signals...)

	return ctx, cancel
}

type cmdFlags struct {
	Config string `flag:"cfg" default:"config.yml"`
}

func loadFlags(cfg interface{}, args []string) error {
	acfg := aconfig.Config{
		SkipFiles:     true,
		SkipEnv:       true,
		SkipDefaults:  true,
		FlagPrefix:    "",
		FlagDelimiter: "-",
		Args:          args,
	}
	loader := aconfig.LoaderFor(cfg, acfg)

	if err := loader.Load(); err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}

	return nil
}
