package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var once sync.Once
var pgPool *pgxpool.Pool

func SetupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	once.Do(func() {
		ctx := context.Background()
		pg, err := postgres.Run(
			ctx,
			"postgres:17-alpine",
			postgres.WithDatabase("test-db"),
			postgres.WithUsername("test-user"),
			postgres.WithPassword("test-pass"),
			testcontainers.WithAdditionalWaitStrategy(
				wait.
					ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second),
			),
		)
		if err != nil {
			t.Errorf("Failed to create the container %s", err)
		}

		connectionStr, err := pg.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			t.Errorf("Cannot get connection string: %s", err)
		}

		pool, err := pgxpool.New(ctx, connectionStr)
		if err != nil {
			t.Errorf("Failed to create pool: %s", err)
		}
		if err := pool.Ping(ctx); err != nil {
			t.Errorf("Pool has not been initialized: %s", err)
		}

		pgPool = pool
	})

	return pgPool
}

func TestInit(t *testing.T) {
	dbPool := SetupDB(t)
	row := dbPool.QueryRow(context.Background(), "select a")
	output := ""

	err := row.Scan(&output)
	if err != nil {
		t.Errorf("Failed to get a row: %s", err)
	}

	t.Log(output)
}
