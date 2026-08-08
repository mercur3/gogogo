package restclient

import (
	"context"
	"errors"
	"fmt"
	"gogogo/internal/otel"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func GetOpenAPI(ctx context.Context) {
	// var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		retry, err := execOpenAPI(
			ctx,
			defaultHTTPClient,
			"https://api.anthropic.com/v1/skills",
			attempt,
		)
		if err == nil {
			break
		}
		// lastErr = err
		if !retry || attempt == maxAttempts {
			break
		}
		if err := backoffSleep(ctx, attempt); err != nil {
			break
		}
	}
}

func execOpenAPI(
	ctx context.Context,
	httpClient *http.Client,
	url string,
	attempt int,
) (bool, error) {
	ctx, span := otel.Tracer().
		Start(ctx, "openapi-rest-client", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attribute.Int("rest-client.request.retry", attempt))
	defer func(t time.Time) {
		span.SetAttributes(
			attribute.Int("rest-client.request.time-ms", int(time.Since(t).Milliseconds())),
		)
		span.End()
	}(time.Now())

	client, err := NewClientWithResponses(url, WithHTTPClient(httpClient))
	if err != nil {
		otel.SetError(span, err)
		return false, fmt.Errorf("failed building the request: %w", err)
	}

	resp, err := client.GetClientWithResponse(
		ctx,
		func(ctx context.Context, req *http.Request) error {
			span.SetAttributes(attribute.String("rest-client.request.url", req.URL.String()))
			span.SetAttributes(attribute.String("rest-client.request.method", req.Method))

			req.Header.Set("Content-Type", jsonOnly)
			req.Header.Set("Accept", jsonOnly)
			req.Header.Set("X-Service-Name", otel.ServiceName)
			return nil
		},
	)
	if err != nil {
		otel.SetError(span, err)

		// network error
		if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
			return true, netErr
		}
		return false, fmt.Errorf("network error: %w", err)
	}
	span.SetAttributes(attribute.Int("rest-client.response.status", resp.StatusCode()))
	span.SetAttributes(
		attribute.Int64("rest-client.response.length", resp.HTTPResponse.ContentLength),
	)

	if isRetryableError(resp.StatusCode()) {
		err := fmt.Errorf("http error=%d", resp.StatusCode())
		otel.SetError(span, err)
		return true, err
	} else if resp.StatusCode() >= 300 {
		return false, fmt.Errorf("http error=%d", resp.StatusCode())
	}

	fmt.Printf("response=%s\n", resp.JSON200)

	return false, nil
}
