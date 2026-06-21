package otel

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel/sdk/metric"
)

type OtelCloser func(context.Context) error
type Closers struct {
	TraceCloser  OtelCloser
	MetricCloser OtelCloser
}

func (c *Closers) CloseResource(ctx context.Context) {
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
