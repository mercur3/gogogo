package middleware

import (
	"context"
	"goweb/internal/api"
	"goweb/internal/common"
	"goweb/internal/otel"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const requestIdHeader string = "X-Request-ID"

type RequestID struct{}

func TraceRequestMiddleware(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		tracer := otel.Tracer()
		ctx, span := tracer.Start(
			ctx,
			"tracer-middleware",
			trace.WithSpanKind(trace.SpanKindServer),
		)

		ctx, requestID := setRequestId(ctx, r, w)
		wrapper := &responseWriterWrapper{statusCode: http.StatusOK, ResponseWriter: w}

		defer func(t time.Time) {
			delta := time.Since(t) / 1000
			span.SetAttributes(attribute.Int("request.time-ms", int(delta)))
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

func OpenApiOperationId(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		ctx, span := otel.Tracer().Start(
			ctx,
			"operation-id-middleware",
			trace.WithSpanKind(trace.SpanKindServer),
		)
		span.SetAttributes(attribute.String("request.operation.id", operationID))
		defer span.End()

		return f(ctx, w, r, request)
	}
}

func MaxRequestBodyMiddleware(
	configs common.Config,
) func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			r.Body = http.MaxBytesReader(w, r.Body, configs.MaxBodySize)

			return f(ctx, w, r, request)
		}
	}
}

func OperationIdMiddleware(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		ctx, span := otel.Tracer().Start(ctx, "operation-id")
		defer span.End()
		span.SetAttributes(attribute.String("request.operation.id", operationID))

		return f(ctx, w, r, request)
	}
}

// setRequestId sets the `X-Request-ID` header empty. It will set a new one if it is not a valid UUID
func setRequestId(
	ctx context.Context,
	r *http.Request,
	w http.ResponseWriter,
) (context.Context, string) {
	var requestID uuid.UUID
	requestIDStr := r.Header.Get(requestIdHeader)

	if requestIDStr == "" {
		requestID = uuid.New()
		requestIDStr = requestID.String()
		r.Header.Set(requestIdHeader, requestIDStr)
	} else {
		parsed, err := uuid.Parse(requestIDStr)
		if err != nil {
			slog.Error("Not a UUID", slog.String(requestIdHeader, requestIDStr))
			parsed = uuid.New()
		}

		requestID = parsed
		requestIDStr = requestID.String()
	}

	w.Header().Set(requestIdHeader, requestIDStr)
	return context.WithValue(ctx, RequestID{}, requestID), requestIDStr
}
