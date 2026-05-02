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
}

func New(dbtx db.DBTX) Repositories {
	q := db.New(dbtx)

	return Repositories{
		Author: &AuthorRepo{q: q},
	}
}

func WithTx[T any](ctx context.Context, pool *pgxpool.Pool, fn func(r *Repositories) (T, error)) (*T, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		slog.Error("failed to create a transaction", slog.Any("error", err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	res, err := fn(new(New(tx)))
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	return &res, nil
}
