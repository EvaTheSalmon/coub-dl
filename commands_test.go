package main

import (
	"path/filepath"
	"testing"
)

func TestExtractPermalink(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"bare permalink", "2uywin", "2uywin"},
		{"full url", "https://coub.com/view/2uywin", "2uywin"},
		{"url without scheme", "coub.com/view/2uywin", "2uywin"},
		{"trailing slash", "https://coub.com/view/2uywin/", "2uywin"},
		{"query string", "https://coub.com/view/2uywin?ref=share", "2uywin"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := extractPermalink(c.in); got != c.want {
				t.Errorf("extractPermalink(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

func TestDateDir(t *testing.T) {
	cases := []struct {
		name    string
		base    string
		updated string
		want    string
	}{
		{"iso timestamp", "videos", "2026-05-15T12:00:00Z", filepath.Join("videos", "2026", "05")},
		{"empty falls back to base", "videos", "", "videos"},
		{"too short falls back to base", "videos", "2026", "videos"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := dateDir(c.base, c.updated); got != c.want {
				t.Errorf("dateDir(%q, %q) = %q, want %q", c.base, c.updated, got, c.want)
			}
		})
	}
}
