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
	errNoCommand      = errors.New("no command provided (serve-bot, server-site, migrate, help)")
	errUnimplemented  = errors.New("unimplemented")
	errUnknownCommand = errors.New("unknown command")
)

func runMain(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return errNoCommand
	}

	var flags cmdFlags
	if err := loadFlags(&flags); err != nil {
		return err
	}

	switch cmd := args[0]; cmd {
	case "serve-bot":
		return serveBotCmd(ctx, flags)
	case "serve-site":
		return serveSiteCmd(ctx, flags)
	case "migrate":
		return migrateCmd(ctx, flags)
	case "help":
		panic(errUnimplemented)
	default:
		return errors.Wrap(errUnknownCommand, cmd)
	}
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

var acfg = aconfig.Config{
	SkipFiles:     true,
	SkipEnv:       true,
	SkipDefaults:  true,
	FlagPrefix:    "",
	FlagDelimiter: "-",
	Args:          os.Args[2:], // Hack to not propagate os.Args to all commands
}

func loadFlags(cfg interface{}) error {
	loader := aconfig.LoaderFor(cfg, acfg)

	if err := loader.Load(); err != nil {
		return fmt.Errorf("cannot load config: %w", err)
	}

	return nil
}
