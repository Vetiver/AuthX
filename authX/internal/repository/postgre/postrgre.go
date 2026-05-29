package postgre

import (
	"authX/utils/config"
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	pool              *pgxpool.Pool
	logger            *zap.Logger
}

func DbStart(cfg *config.Config, logger *zap.Logger) *pgxpool.Pool {
	connStr := cfg.DatabaseURL
	logger.Info("Service init: connecting to database...")
	urlExample := string(connStr)
	dbpool, err := pgxpool.Connect(context.Background(), urlExample)
	if err != nil {
		logger.Error("db pool connect error", zap.Error(err))
		return nil
	}

	logger.Info("db connection successful")
	return dbpool
}

func NewRepository(pool *pgxpool.Pool, logger *zap.Logger) *DB {
	db := &DB{
		pool:              pool,
		logger:            logger,
	}

	err := db.Migrate(context.Background())
	if err != nil {
		db.logger.Error("db migration error", zap.Error(err))
		return nil
	}

	return db
}
