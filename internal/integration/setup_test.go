package integration

import (
	"context"
	"goweb/internal/db"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var pgPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()
	pg, err := postgres.Run(
		ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("test-user"),
		postgres.WithPassword("test-pass"),
		testcontainers.WithReuseByName("my-test-postgres"),
		testcontainers.WithAdditionalWaitStrategy(
			wait.
				ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	assertNoError(err)

	connectionStr, err := pg.ConnectionString(ctx, "sslmode=disable")
	assertNoError(err)

	pgPool, err = pgxpool.New(ctx, connectionStr)
	assertNoError(err)
	assertNoError(pgPool.Ping(ctx))

	// run the migrations
	assertNoError(db.RunMigrations(pgPool))

	code := m.Run() // run ALL tests in this package once

	pgPool.Close()
	assertNoError(testcontainers.TerminateContainer(pg))
	os.Exit(code)
}

func assertNoError(err error) {
	if err != nil {
		panic(err)
	}
}
