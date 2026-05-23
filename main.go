package main

//go:generate go tool oapi-codegen -config ./assets/api/cfg.yaml ./assets/api/api.yaml
//go:generate go tool sqlc generate -f ./assets/sqlc.yaml

import (
	"context"
	"errors"
	"goweb/internal/db"
	"goweb/internal/handle"
	"goweb/internal/otel"
	"goweb/internal/repository"
	"goweb/internal/service"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	configureSlog()

	// make signal channel
	sigCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	pool, err := db.InitPool(sigCtx, "exam-user", "exam-password", "exam-db", "5432")
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	otelClosers, err := otel.InitOtel(sigCtx)
	defer closeOtel(sigCtx, otelClosers)
	if err != nil {
		log.Fatal(err)
	}

	repositories := repository.New(pool)
	author := service.AuthorService(&repositories)

	// srv := handle.MakeServer(author)
	srv := handle.MakeServerFromOpenAPI(author)
	go func() {
		slog.Info("Server starting")

		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("Server is closing")
		} else if err != nil {
			slog.Error("Server error", slog.Any("error", err))
		}
	}()
	go func() {
		log.Println("pprof listening on localhost:6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Println("pprof server error:", err)
		}
	}()

	// wait for signal
	<-sigCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	slog.Info("received shutdown signal")

	// gracefull shutdown
	var wg sync.WaitGroup

	// close the server
	wg.Go(func() {
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("error during the shutdown", slog.Any("error", err))

			if err := srv.Close(); err != nil {
				slog.Error("failed to close the server with force", slog.Any("error", err))
			}
		}
	})

	// close otel
	wg.Go(func() {
		closeOtel(ctx, otelClosers)
	})

	// close db
	wg.Go(func() {
		slog.Info("Closing the DB")
		pool.Close()
	})

	wg.Wait()

	slog.Info("everything was closed")
}

func closeOtel(ctx context.Context, c otel.Closers) {
	slog.Info("closing tracer")
	if err := c.TraceCloser(ctx); err != nil {
		slog.Error("cannot close tracer", slog.Any("error", err))
	}

	slog.Info("closing meter")
	if err := c.MetricCloser(ctx); err != nil {
		if !errors.Is(err, metric.ErrReaderShutdown) {
			slog.Error("cannot close meter", slog.Any("error", err))
		}
	}
}

func configureSlog() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
}
