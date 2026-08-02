package otel

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel/sdk/metric"
)

type OtelCloser func(context.Context) error
type Closers struct {
	TraceCloser  OtelCloser
	MetricCloser OtelCloser
	LogCloser    OtelCloser
}

func (c *Closers) CloseResource(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Go(func() {
		slog.InfoContext(ctx, "closing tracer")
		if err := c.TraceCloser(ctx); err != nil {
			slog.ErrorContext(ctx, "cannot close tracer", slog.Any("error", err))
		}
	})
	wg.Go(func() {
		slog.InfoContext(ctx, "closing meter")
		if err := c.MetricCloser(ctx); err != nil {
			if !errors.Is(err, metric.ErrReaderShutdown) {
				slog.ErrorContext(ctx, "cannot close meter", slog.Any("error", err))
			}
		}
	})
	wg.Go(func() {
		slog.InfoContext(ctx, "clossing logger")
		if err := c.LogCloser(ctx); err != nil {
			slog.ErrorContext(ctx, "cannot close logger", slog.Any("error", err))
		}
	})

	wg.Wait()
}
