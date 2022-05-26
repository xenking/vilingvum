package main

import (
	"context"
	"os"
	"time"

	"github.com/cloudflare/tableflip"
	"github.com/phuslu/log"

	"github.com/xenking/vilingvum/config"
	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/internal/application"
	"github.com/xenking/vilingvum/pkg/logger"
)

func serveBotCmd(ctx context.Context, flags cmdFlags) error {
	cfg, err := config.NewConfig(flags.Config)
	if err != nil {
		return err
	}
	log.Debug().Msgf("%+v", cfg)

	l := logger.New(cfg.Log)
	logger.SetGlobal(l)
	log.DefaultLogger = *logger.NewModule("global")

	// Check if migration needed
	if cfg.MigrationMode {
		err = migrateDatabase(cfg)
		if err != nil {
			return err
		}
	}

	db, err := database.Init(ctx, cfg.Postgres)
	if err != nil {
		return err
	}

	bot, err := application.New(ctx, cfg.Bot, db)
	if err != nil {
		return err
	}

	return serveBot(ctx, cfg, bot)
}

func serveBot(ctx context.Context, cfg *config.Config, bot *application.Bot) error {
	upg, listerErr := tableflip.New(tableflip.Options{
		UpgradeTimeout: cfg.GracefulShutdownDelay,
	})
	if listerErr != nil {
		return listerErr
	}
	defer upg.Stop()

	// waiting for ctrl+c
	go func() {
		<-ctx.Done()
		upg.Stop()
	}()

	go bot.Bot.Start()

	log.Info().Msg("service ready")
	if upgErr := upg.Ready(); upgErr != nil {
		return upgErr
	}

	<-upg.Exit()
	log.Info().Msg("shutting down")

	time.AfterFunc(cfg.GracefulShutdownDelay, func() {
		log.Warn().Msg("graceful shutdown timed out")
		os.Exit(1)
	})

	bot.Stop()

	return listerErr
}
