package database

import (
	"context"

	"github.com/jackc/pgtype"
	pgtypeuuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/xenking/vilingvum/config"
	"github.com/xenking/vilingvum/database/adapter"
	"github.com/xenking/vilingvum/pkg/logger"
)

type DB struct {
	*Queries
	pool *pgxpool.Pool
}

func Init(ctx context.Context, cfg config.PostgresConfig) (*DB, error) {
	databaseCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}
	databaseCfg.ConnConfig.Logger = adapter.NewPgLogger(logger.NewModule("pg"))
	ll, err := pgx.LogLevelFromString(cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	databaseCfg.ConnConfig.LogLevel = ll
	databaseCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		conn.ConnInfo().RegisterDataType(pgtype.DataType{
			Value: &pgtypeuuid.UUID{},
			Name:  "uuid",
			OID:   pgtype.UUIDOID,
		})

		return nil
	}

	pool, err := pgxpool.ConnectConfig(ctx, databaseCfg)
	if err != nil {
		return nil, err
	}

	return &DB{
		Queries: New(pool),
		pool:    pool,
	}, nil
}
