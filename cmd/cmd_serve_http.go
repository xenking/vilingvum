package main

import (
	"context"
	"os"
	"time"

	"github.com/cloudflare/tableflip"
	"github.com/phuslu/log"

	"github.com/xenking/vilingvum/api/server"
	"github.com/xenking/vilingvum/config"
	"github.com/xenking/vilingvum/database"
	"github.com/xenking/vilingvum/pkg/logger"
)

func serveHTTPCmd(ctx context.Context, flags cmdFlags) error {
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

	return serveHTTP(ctx, cfg, db)
}

func serveHTTP(ctx context.Context, cfg *config.Config, db *database.DB) error {
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

	srv := server.New(cfg.Server, db)

	// Serve must be called before Ready
	srvListener, listenErr := upg.Listen("tcp", cfg.Server.Addr)
	if listenErr != nil {
		log.Error().Err(listenErr).Msg("can't listen")

		return listenErr
	}

	// run gateway server
	go func() {
		if serveErr := srv.Serve(srvListener); serveErr != nil {
			log.Error().Err(serveErr).Msg("site server")
		}
	}()
	log.Info().Msg("serving site server")

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

	if srv != nil {
		listerErr = srv.Shutdown()
	}

	return listerErr
}
