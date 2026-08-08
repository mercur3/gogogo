package restclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// constantSleep returns a sleepFunc that waits d (or returns ctx.Err early)
// regardless of attempt number. Deterministic and fast for tests — no
// package state involved, so it's safe under t.Parallel().
func constantSleep(d time.Duration) sleepFunc {
	return func(ctx context.Context, _ int) error {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-timer.C:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

const testBackoff = 2 * time.Millisecond

func TestGet_SucceedsFirstTry(t *testing.T) {
	t.Parallel()

	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)

		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, jsonOnly, r.Header.Get("Accept"))
		assert.NotEmpty(t, r.Header.Get("X-Service-Name"))

		w.Header().Set("Content-Type", jsonOnly)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		assert.NoError(t, err)
	}))
	defer srv.Close()

	err := get(context.Background(), srv.URL, constantSleep(testBackoff))

	assert.NoError(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&requests))
}

func TestGet_RetriesOnRetryableStatus_ThenSucceeds(t *testing.T) {
	t.Parallel()

	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&requests, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"status":"ok"}`))
		assert.NoError(t, err)
	}))
	defer srv.Close()

	err := get(context.Background(), srv.URL, constantSleep(testBackoff))

	assert.NoError(t, err)
	assert.EqualValues(t, 3, atomic.LoadInt32(&requests))
}

func TestGet_ExhaustsRetriesOnPersistentRetryableStatus(t *testing.T) {
	t.Parallel()

	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	err := get(context.Background(), srv.URL, constantSleep(testBackoff))

	assert.Error(t, err)
	assert.EqualValues(t, maxAttempts, atomic.LoadInt32(&requests))
}

func TestGet_DoesNotRetryOnNonRetryableStatus(t *testing.T) {
	t.Parallel()

	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	err := get(context.Background(), srv.URL, constantSleep(testBackoff))

	assert.Error(t, err)
	assert.EqualValues(t, 1, atomic.LoadInt32(&requests))
}

func TestGet_ResponseBodyTooLarge(t *testing.T) {
	t.Parallel()

	oversized := make([]byte, maxResponseSize+100)
	for i := range oversized {
		oversized[i] = 'a'
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(oversized)
		assert.NoError(t, err)
	}))
	defer srv.Close()

	err := get(context.Background(), srv.URL, constantSleep(testBackoff))

	assert.Error(t, err)
}

func TestGet_StopsRetryingWhenContextCancelled(t *testing.T) {
	t.Parallel()

	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	// Long enough to reliably cancel mid-backoff rather than race the timer.
	longBackoff := constantSleep(300 * time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := get(ctx, srv.URL, longBackoff)
	elapsed := time.Since(start)

	assert.Error(t, err)
	// Current implementation returns the last HTTP error, not
	// context.Canceled, on this path (see review) — timing only here.
	assert.Less(t, elapsed, 200*time.Millisecond)
	assert.EqualValues(t, 1, atomic.LoadInt32(&requests))
}
