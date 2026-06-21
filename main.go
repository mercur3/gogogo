package main

//go:generate go tool oapi-codegen -config ./assets/api/cfg.yaml ./assets/api/api.yaml
//go:generate go tool sqlc generate -f ./assets/sqlc.yaml

import (
	"context"
	"errors"
	"goweb/internal/common"
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
)

func main() {
	configureSlog()
	cfg, err := common.ParseConfigs()
	if err != nil {
		panic(err)
	}

	// make signal channel
	sigCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	otelCloser, err := otel.InitOtel(sigCtx, "localhost:4317")
	defer otelCloser.CloseResource(sigCtx)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := db.InitPool(sigCtx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repositories := new(repository.New(pool.Pool))
	author := service.AuthorService(repositories)
	book := service.BookService(repositories)

	// srv := handle.MakeServer(author)
	srv := handle.MakeServerFromOpenAPI(cfg, author, book)
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
	closeResources(ctx, &otelCloser, srv, pool)
}

func configureSlog() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(handler))
}

func closeResources(ctx context.Context, resources ...common.ResourceCloser) {
	var wg sync.WaitGroup

	for _, r := range resources {
		wg.Go(func() { r.CloseResource(ctx) })
	}

	wg.Wait()
	slog.Info("everything was closed")
}
