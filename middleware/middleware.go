package middleware

import (
	"goweb/otel"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func TimedRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tracer := otel.Tracer()
		ctx, span := tracer.Start(r.Context(), "timed-middleware")

		defer func(t time.Time) {
			delta := time.Since(t) / 1000
			span.SetAttributes(attribute.Int("request.time", int(delta)))
			span.SetAttributes(attribute.String("url.path", r.URL.Path))
			span.SetAttributes(attribute.String("url.method", r.Method))
			span.SetAttributes(attribute.String("url.host", r.Host))

			span.End()
		}(time.Now())

		h.ServeHTTP(w, r.WithContext(ctx))
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
