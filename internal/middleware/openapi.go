package middleware

import (
	"context"
	"goweb/internal/api"
	"goweb/internal/otel"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

const RequestID string = "X-Request-ID"

func TraceRequestMiddleware(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		tracer := otel.Tracer()
		ctx, span := tracer.Start(r.Context(), "tracer-middleware")

		wrapper := new(responseWriterWrapper{statusCode: http.StatusOK, ResponseWriter: w})

		requestID := r.Header.Get(RequestID)
		if requestID == "" {
			requestID = uuid.NewString()
			r.Header.Set(RequestID, requestID)
		}

		defer func(t time.Time) {
			delta := time.Since(t) / 1000
			span.SetAttributes(attribute.Int("request.time", int(delta)))
			span.SetAttributes(attribute.String("request.id", requestID))
			span.SetAttributes(attribute.String("request.operation.id", operationID))
			span.SetAttributes(attribute.Int("request.http.status", wrapper.statusCode))

			span.SetAttributes(attribute.String("url.path", r.URL.Path))
			span.SetAttributes(attribute.String("url.method", r.Method))
			span.SetAttributes(attribute.String("url.host", r.Host))

			span.End()
		}(time.Now())

		return f(ctx, w, r, request)
	}
}

func MaxRequestBody(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		return f(ctx, w, r, request)
	}
}
