package restclient

import (
	"context"
	"errors"
	"fmt"
	"gogogo/internal/otel"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func Get(ctx context.Context) error {
	return get(ctx, defaultURL, backoffSleep)
}

func get(ctx context.Context, url string, sleep sleepFunc) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		retry, err := exec(ctx, url, http.MethodGet, attempt)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retry || attempt == maxAttempts {
			break
		}
		if err := sleep(ctx, attempt); err != nil {
			break
		}
	}
	return lastErr
}

// TODO fix this
// returns true if can retry
func exec(ctx context.Context, url string, method string, attempt int) (bool, error) {
	ctx, span := otel.Tracer().Start(ctx, "rest-client", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(attribute.Int("rest-client.request.retry", attempt))
	span.SetAttributes(attribute.String("rest-client.request.url", url))
	span.SetAttributes(attribute.String("rest-client.request.method", method))
	defer func(t time.Time) {
		span.SetAttributes(
			attribute.Int("rest-client.request.time-ms", int(time.Since(t).Milliseconds())),
		)
		span.End()
	}(time.Now())

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		otel.SetError(span, err)
		return false, fmt.Errorf("failed building the request: %w", err)
	}
	req.Header.Set("Content-Type", jsonOnly)
	req.Header.Set("Accept", jsonOnly)
	req.Header.Set("X-Service-Name", otel.ServiceName)

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		otel.SetError(span, err)

		// network error
		if netErr, ok := errors.AsType[net.Error](err); ok && netErr.Timeout() {
			return true, netErr
		}
		return false, fmt.Errorf("network error: %w", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			slog.Error("failed closing the response", slog.String("error", err.Error()))
		}
	}()
	span.SetAttributes(attribute.Int("rest-client.response.status", resp.StatusCode))
	span.SetAttributes(attribute.Int64("rest-client.response.length", resp.ContentLength))

	if isRetryableError(resp.StatusCode) {
		// drain the connection or else it might initiate another TCP handshake instead of reusing the connection
		_, err := io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseSize))
		if err != nil {
			slog.Error("failed draining response body", slog.String("error", err.Error()))
		}

		err = fmt.Errorf("http error=%d", resp.StatusCode)
		otel.SetError(span, err)
		return true, err
	} else if resp.StatusCode >= 300 {
		return false, fmt.Errorf("http error=%d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(http.MaxBytesReader(nil, resp.Body, maxResponseSize))
	if err != nil {
		otel.SetError(span, err)
		return false, fmt.Errorf("failed reading response: %w", err)
	}
	fmt.Printf("response=%s\n", respBody)

	return false, nil
}
