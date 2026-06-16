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
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		base := time.Duration(1<<(attempt-1)) * time.Second
		low, high := base/2, base*3/2
		for range 1000 {
			d := backoff(attempt)
			if d < low || d >= high {
				t.Fatalf("backoff(%d) = %v, want within [%v, %v)", attempt, d, low, high)
			}
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want time.Duration
	}{
		{"empty", "", 0},
		{"zero seconds", "0", 0},
		{"positive seconds", "5", 5 * time.Second},
		{"large seconds", "120", 120 * time.Second},
		{"negative seconds", "-3", 0},
		{"garbage", "soon", 0},
		{"past http date", "Mon, 02 Jan 2006 15:04:05 GMT", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := parseRetryAfter(c.in); got != c.want {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", c.in, got, c.want)
			}
		})
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
