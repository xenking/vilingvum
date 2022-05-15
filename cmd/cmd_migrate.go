package main

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/phuslu/log"
	"github.com/pkg/errors"

	"tgbot/config"
	"tgbot/database"
	"tgbot/pkg/logger"
)

func migrateCmd(ctx context.Context, flags cmdFlags) error {
	cfg, err := config.NewConfig(flags.Config)
	if err != nil {
		return err
	}

	l := logger.New(cfg.Log)
	logger.SetGlobal(l)
	log.DefaultLogger = *logger.NewModule("global")

	_, err = database.Init(ctx, cfg.Postgres)
	if err != nil {
		return err
	}

	return migrateDatabase(cfg)
}

func migrateDatabase(cfg *config.Config) error {
	p := &pgx.Postgres{}
	d, err := p.Open(cfg.Postgres.DSN)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := d.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("close pgx connection")
		}
	}()
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+cfg.Postgres.MigrationsDir,
		"pgx", d)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}

	return nil
}
