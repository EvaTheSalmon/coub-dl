package coub

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const maxAttempts = 3

func (c *Client) doGet(ctx context.Context, url string) (*http.Response, error) {
	for attempt := 1; ; attempt++ {
		resp, retryable, err := c.try(ctx, url)
		if err == nil {
			return resp, nil
		}

		if attempt >= maxAttempts || !retryable {
			return nil, err
		}

		if err := wait(ctx, backoff(attempt)); err != nil {
			return nil, err
		}
	}
}

func (c *Client) try(ctx context.Context, url string) (*http.Response, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("creating request: %s", redact(err.Error(), c.token))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ctx.Err() == nil, fmt.Errorf("executing request: %s", redact(err.Error(), c.token))
	}

	if resp.StatusCode == http.StatusOK {
		return resp, false, nil
	}

	resp.Body.Close()
	retryable := resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
	return nil, retryable, fmt.Errorf("unexpected status %d", resp.StatusCode)
}

func backoff(attempt int) time.Duration {
	return time.Duration(1<<(attempt-1)) * time.Second
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
