package coub

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	maxAttempts   = 3
	maxRetryAfter = 30 * time.Second
)

func (c *Client) doGet(ctx context.Context, url string) (*http.Response, error) {
	for attempt := 1; ; attempt++ {
		resp, retryable, retryAfter, err := c.try(ctx, url)
		if err == nil {
			return resp, nil
		}

		if attempt >= maxAttempts || !retryable {
			return nil, err
		}

		delay := backoff(attempt)
		if retryAfter > 0 {
			delay = min(retryAfter, maxRetryAfter)
		}

		if err := wait(ctx, delay); err != nil {
			return nil, err
		}
	}
}

func (c *Client) try(ctx context.Context, url string) (*http.Response, bool, time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, 0, fmt.Errorf("creating request: %s", redact(err.Error(), c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ctx.Err() == nil, 0, fmt.Errorf("executing request: %s", redact(err.Error(), c.token))
	}

	if resp.StatusCode == http.StatusOK {
		return resp, false, 0, nil
	}

	retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
	resp.Body.Close()
	retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
	return nil, retryable, retryAfter, fmt.Errorf("unexpected status %d", resp.StatusCode)
}

func backoff(attempt int) time.Duration {
	base := time.Duration(1<<(attempt-1)) * time.Second
	return time.Duration(float64(base) * (0.5 + rand.Float64()))
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	if secs, err := strconv.Atoi(header); err == nil && secs > 0 {
		return time.Duration(secs) * time.Second
	}

	date, err := http.ParseTime(header)
	if err != nil {
		return 0
	}
	return max(time.Until(date), 0)
}

func wait(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-t.C:
		return nil
	}
}
