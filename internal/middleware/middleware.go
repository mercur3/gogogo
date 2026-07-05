package middleware

import (
	"goweb/internal/otel"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type responseWriterWrapper struct {
	statusCode int
	http.ResponseWriter
}

func (r *responseWriterWrapper) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func TraceRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer()
		ctx, span := tracer.Start(
			r.Context(),
			"timed-middleware",
			trace.WithSpanKind(trace.SpanKindServer),
		)

		ctx, requestID := setRequestId(ctx, r, w)
		wrapper := &responseWriterWrapper{statusCode: http.StatusOK, ResponseWriter: w}

		defer func(t time.Time) {
			delta := time.Since(t) / 1000
			span.SetAttributes(attribute.Int("request.time-ms", int(delta)))
			span.SetAttributes(attribute.String("request.id", requestID))
			span.SetAttributes(attribute.Int("request.http.status", wrapper.statusCode))

			span.SetAttributes(attribute.String("url.path", r.URL.Path))
			span.SetAttributes(attribute.String("url.method", r.Method))
			span.SetAttributes(attribute.String("url.host", r.Host))

			span.End()

			span.End()
		}(time.Now())

		h.ServeHTTP(wrapper, r.WithContext(ctx))
	})
}

func DeprecatedEndpoint(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := otel.Tracer().Start(r.Context(), "deprecated-endpoint")
		defer span.End()
		span.SetAttributes(attribute.Bool("endpoint.deprecated", true))

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
