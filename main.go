package main

import (
	"context"
	"errors"
	"fmt"
	"goweb/middleware"
	"goweb/otel"
	"log"
	"log/slog"
	"net/http"
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

	otelClosers, err := otel.InitOtel(sigCtx)
	defer closeOtel(sigCtx, otelClosers)
	if err != nil {
		log.Fatal(err)
	}

	srv := makeServer()
	go func() {
		slog.Info("Server starting")

		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("Server is closing")
		} else if err != nil {
			slog.Error("Server error", slog.Any("error", err))
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

	wg.Wait()

	slog.Info("everything was closed")
}

func makeServer() *http.Server {
	v1 := http.NewServeMux()
	v1.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeBody(w, "GET /v1/test")
	})
	v1.HandleFunc("POST /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		writeBody(w, "POST /v1/test")
	})
	v1.HandleFunc("GET /test/{id}", func(w http.ResponseWriter, r *http.Request) {
		val := r.PathValue("id")
		if val != "" {
			w.WriteHeader(http.StatusOK)
			writeBody(w, fmt.Sprintf("GET /v1/test/{%s}", val))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	v2 := http.NewServeMux()
	v2.HandleFunc("GET /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		writeBody(w, "GET /v2/test")
	})
	v2.HandleFunc("POST /test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		writeBody(w, "POST /v2/test")
	})
	v2.HandleFunc("GET /test/{id}", func(w http.ResponseWriter, r *http.Request) {
		val := r.PathValue("id")
		if val != "" {
			w.WriteHeader(http.StatusOK)
			writeBody(w, fmt.Sprintf("GET /v2/test/{%s}", val))
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})

	mux := http.NewServeMux()
	mux.Handle("/v1/", middleware.DeprecatedEndpoint(http.StripPrefix("/v1", v1)))
	mux.Handle("/v2/", http.StripPrefix("/v2", v2))

	return &http.Server{
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      middleware.TimedRequest(mux),
	}
}

func writeBody(w http.ResponseWriter, s string) {
	_, err := w.Write([]byte(s))
	if err != nil {
		slog.Error("failed to write body", slog.Any("error", err))
	}
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
