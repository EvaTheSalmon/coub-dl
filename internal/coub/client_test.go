package coub

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestClient(srv *httptest.Server) *Client {
	return NewClient(srv.Client(), "")
}

func TestRedact(t *testing.T) {
	cases := []struct {
		name   string
		msg    string
		secret string
		want   string
	}{
		{"replaces secret", "GET ...?api_token=abc123 failed", "abc123", "GET ...?api_token=*** failed"},
		{"empty secret is no-op", "nothing to hide", "", "nothing to hide"},
		{"secret not present", "clean message", "abc123", "clean message"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := redact(c.msg, c.secret); got != c.want {
				t.Errorf("redact(%q, %q) = %q, want %q", c.msg, c.secret, got, c.want)
			}
		})
	}
}

func TestGetJSONDecodes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"permalink":"2uywin","title":"hi","tags":[{"title":"cat"}]}`))
	}))
	defer srv.Close()

	var coub Coub
	if err := newTestClient(srv).getJSON(context.Background(), srv.URL, &coub); err != nil {
		t.Fatalf("getJSON: %v", err)
	}

	if coub.Permalink != "2uywin" {
		t.Errorf("Permalink = %q, want %q", coub.Permalink, "2uywin")
	}
	if coub.Title != "hi" {
		t.Errorf("Title = %q, want %q", coub.Title, "hi")
	}
	if len(coub.Tags) != 1 || coub.Tags[0].Title != "cat" {
		t.Errorf("Tags = %+v, want one tag %q", coub.Tags, "cat")
	}
}
