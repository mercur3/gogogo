package integration

import (
	"context"
	"errors"
	"goweb/internal/db"
	"goweb/internal/repository"
	"testing"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestDbInitialization(t *testing.T) {
	t.Parallel()

	row := pgPool.QueryRow(context.Background(), "select 1")
	output := ""

	err := row.Scan(&output)
	assert.NoError(t, err)
	assert.Equal(t, "1", output)
}

func Test_WithTx_will_not_commit_if_err(t *testing.T) {
	t.Parallel()

	repos := repository.New(pgPool)
	out, err := repository.WithTx(
		context.Background(),
		&repos,
		func(ctx context.Context) (*db.Book, error) {
			book, err := repos.Book.CreateBook(ctx, db.CreateBookParams{
				Title:       "test",
				PublishedAt: time.Date(-5000, time.April, 1, 0, 0, 0, 0, time.UTC),
			})

			return &book, err
		},
	)

	assert.Nil(t, out)

	pgErr, bool := errors.AsType[*pgconn.PgError](err)
	assert.True(t, bool)
	assert.Equal(t, pgerrcode.DatetimeFieldOverflow, pgErr.Code)
}
