package restclient

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const jsonOnly = "application/json; charset=utf-8"
const maxResponseSize = 1024
const maxAttempts = 3
const defaultURL = "https://api.uber.com/v1.2/products?latitude=37.7752315&longitude=-122.418075"

var defaultHTTPClient = makeHTTPClient()

// sleepFunc waits before the next retry attempt, returning early if ctx
// ends. It's a parameter rather than a package var so callers — including
// tests — can supply their own without mutating shared state. That's what
// makes get() safe to exercise under t.Parallel().
type sleepFunc func(ctx context.Context, attempt int) error

func makeHTTPClient() *http.Client {
	transport := http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		DisableKeepAlives:   false,
		// TLSClientConfig: &tls.Config{
		// 	MinVersion: tls.VersionTLS12,
		// },
	}

	return &http.Client{
		Transport: &transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			maxRedirects := 10
			if len(via) >= maxRedirects {
				return fmt.Errorf("exceeded the max number of redirect=%d", maxRedirects)
			}

			return nil
		},
	}
}

func isRetryableError(code int) bool {
	return code == http.StatusTooManyRequests || code == http.StatusBadGateway ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusGatewayTimeout
}

// backoffSleep is the production sleepFunc: exponential backoff with
// [-15%, +15%] jitter, one second per unit of 2^attempt.
func backoffSleep(ctx context.Context, attempt int) error {
	exponential := 1 << attempt
	multiplier := rand.Intn(31)
	factor := float64(exponential) + float64((multiplier-15)*exponential)/100

	timer := time.NewTimer(time.Duration(factor * float64(time.Second)))
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
