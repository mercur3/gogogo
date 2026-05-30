package repository

import (
	"context"
	"fmt"
	"goweb/internal/db"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repositories struct {
	Author *AuthorRepo
	Book   *BookRepo
	pool   *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Repositories {
	q := db.New(pool)

	return Repositories{
		Author: &AuthorRepo{q},
		Book:   &BookRepo{q},
		pool:   pool,
	}
}

func WithTx[T any](
	ctx context.Context,
	r *Repositories,
	fn func(ctx context.Context) (T, error),
) (*T, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		slog.Error("failed to create a transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // rollback after commit is a harmless no-op

	res, err := fn(ctx)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &res, nil
}
