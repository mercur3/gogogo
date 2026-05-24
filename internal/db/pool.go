package db

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"runtime"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func InitPool(
	ctx context.Context,
	username string,
	password string,
	dbName string,
	port string,
) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(
		fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", username, password, port, dbName),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the db: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the db: %w", err)
	}

	slog.Info("DB pool has been created")
	return pool, RunMigrations(pool)
}

func RunMigrations(pool *pgxpool.Pool) error {
	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return err
	}

	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close the db connection after migration", slog.Any("error", err))
		}
	}()

	slog.Info("Running the migrations")

	_, filepath, _, _ := runtime.Caller(0)
	dirpath := path.Join(filepath, "..", "..", "..", "assets", "migrations")
	if err := goose.Up(db, dirpath); err != nil {
		return fmt.Errorf("failed to run the migrations: %w", err)
	}

	return nil
}
