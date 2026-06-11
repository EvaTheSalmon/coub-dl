package coub

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
	}
	for _, c := range cases {
		if got := backoff(c.attempt); got != c.want {
			t.Errorf("backoff(%d) = %v, want %v", c.attempt, got, c.want)
		}
	}
}

func TestDoGetNoRetryOn404(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	if _, err := newTestClient(srv).doGet(context.Background(), srv.URL); err == nil {
		t.Fatal("expected an error on 404")
	}
	if n := atomic.LoadInt32(&calls); n != 1 {
		t.Errorf("404 must not be retried: got %d calls, want 1", n)
	}
}

func TestDoGetRetriesUntilExhausted(t *testing.T) {
	if testing.Short() {
		t.Skip("retry backoff sleeps real seconds")
	}
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	if _, err := newTestClient(srv).doGet(context.Background(), srv.URL); err == nil {
		t.Fatal("expected an error after retries are exhausted")
	}
	if n := atomic.LoadInt32(&calls); n != maxAttempts {
		t.Errorf("503 should retry up to the limit: got %d calls, want %d", n, maxAttempts)
	}
}

func TestDoGetCancelledContextStops(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := newTestClient(srv).doGet(ctx, srv.URL); err == nil {
		t.Fatal("expected an error with a cancelled context")
	}
	if n := atomic.LoadInt32(&calls); n > 1 {
		t.Errorf("cancelled context must not keep retrying: got %d calls", n)
	}
}
