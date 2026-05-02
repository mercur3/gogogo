package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitPool(ctx context.Context, username string, password string, dbName string, port string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", username, password, port, dbName))
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the db")
	}

	slog.Info("DB pool has been created")
	return pool, nil
}
