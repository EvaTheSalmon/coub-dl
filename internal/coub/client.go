package coub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	apiBase     = "https://coub.com/api/v2"
	perPage     = 25
	likesBuffer = 50
)

type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(hc *http.Client, token string) *Client {
	return &Client{httpClient: hc, token: token}
}

func (c *Client) Get(ctx context.Context, permalink string) (Coub, error) {
	url := fmt.Sprintf("%s/coubs/%s", apiBase, permalink)

	var coub Coub
	if err := c.getJSON(ctx, url, &coub); err != nil {
		return Coub{}, err
	}
	return coub, nil
}

func (c *Client) FetchLikes(ctx context.Context) (<-chan Coub, <-chan error) {
	ch := make(chan Coub, likesBuffer)
	errCh := make(chan error, 1)

	go func() {
		defer close(ch)
		defer close(errCh)

		for page := 1; ; page++ {
			resp, err := c.fetchLikesPage(ctx, page)
			if err != nil {
				errCh <- fmt.Errorf("fetching page %d: %w", page, err)
				return
			}

			for _, coub := range resp.Coubs {
				select {
				case <-ctx.Done():
					return
				case ch <- coub:
				}
			}

			if page >= resp.TotalPages {
				return
			}
		}
	}()

	return ch, errCh
}

func (c *Client) fetchLikesPage(ctx context.Context, page int) (CoubResponse, error) {
	url := fmt.Sprintf("%s/timeline/likes?page=%d&per_page=%d&api_token=%s", apiBase, page, perPage, c.token)

	var resp CoubResponse
	if err := c.getJSON(ctx, url, &resp); err != nil {
		return CoubResponse{}, err
	}
	return resp, nil
}

func (c *Client) LikesCount(ctx context.Context) (int, error) {
	resp, err := c.fetchLikesPage(ctx, 1)
	if err != nil {
		return 0, err
	}
	return resp.TotalPages * perPage, nil
}

func (c *Client) getJSON(ctx context.Context, url string, dst any) error {
	resp, err := c.doGet(ctx, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return fmt.Errorf("decoding: %w", err)
	}
	return nil
}

func redact(msg, secret string) string {
	if secret == "" {
		return msg
	}
	return strings.ReplaceAll(msg, secret, "***")
}
