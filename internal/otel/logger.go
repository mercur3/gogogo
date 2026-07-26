package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/global"
)

// WrapSlogHandler returns a slog.Handler that writes to base (e.g. your
// existing stdout JSON handler) AND forwards every record to the OTel
// LoggerProvider, so records keep going to the terminal/container logs for
// local dev while also being shipped to the collector and on to Loki, with
// trace/span IDs attached automatically when the record's context carries
// an active span.
//
// Call this only after InitOtel has run, since it reads the currently
// registered global LoggerProvider.
func WrapSlogHandler(base slog.Handler, serviceName string) slog.Handler {
	otelHandler := otelslog.NewHandler(
		serviceName,
		otelslog.WithLoggerProvider(global.GetLoggerProvider()),
	)
	return &multiHandler{handlers: []slog.Handler{base, otelHandler}}
}

// multiHandler fans a slog.Record out to several handlers. It's a small,
// dependency-free stand-in for what libraries like samber/slog-multi do.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		if err := h.Handle(ctx, r.Clone()); err != nil {
			return err
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: next}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: next}
}
