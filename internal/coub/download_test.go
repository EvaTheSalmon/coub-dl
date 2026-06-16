package coub

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestBestURL(t *testing.T) {
	cases := []struct {
		name     string
		variants []MediaVariant
		want     string
	}{
		{"first non-empty wins", []MediaVariant{{URL: "higher"}, {URL: "high"}}, "higher"},
		{"falls through empties", []MediaVariant{{URL: ""}, {URL: "high"}, {URL: "med"}}, "high"},
		{"all empty", []MediaVariant{{URL: ""}, {URL: ""}}, ""},
		{"nil slice", nil, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := bestURL(c.variants); got != c.want {
				t.Errorf("bestURL(%v) = %q, want %q", c.variants, got, c.want)
			}
		})
	}
}

func TestFilenameSafe(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"typical permalink", "2uywin", true},
		{"alnum and underscore", "Abc_123", true},
		{"inner dash allowed", "a-b-c", true},
		{"leading underscore", "_foo", true},

		{"empty", "", false},
		{"leading dash looks like flag", "-x", false},
		{"dot disallowed", "a.b", false},
		{"parent traversal", "..", false},
		{"relative traversal", "../../tmp/evil", false},
		{"forward slash", "foo/bar", false},
		{"backslash", `foo\bar`, false},
		{"space", "foo bar", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := filenameSafe(c.in); got != c.want {
				t.Errorf("filenameSafe(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestUserNameSafe(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"plain name", "my clip", true},
		{"spaces allowed", "hello world", true},
		{"inner dot allowed", "clip.v2", true},
		{"unicode allowed", "клип", true},
		{"leading dot ok", ".hidden", true},

		{"empty", "", false},
		{"leading dash looks like flag", "-i", false},
		{"forward slash", "sub/clip", false},
		{"backslash", `sub\clip`, false},
		{"parent traversal", "..", false},
		{"relative traversal", "../evil", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := userNameSafe(c.in); got != c.want {
				t.Errorf("userNameSafe(%q) = %v, want %v", c.in, got, c.want)
			}
		})
	}
}

func TestDownloadRejectsUnsafeNames(t *testing.T) {
	cases := []struct {
		label     string
		permalink string
		name      string
		wantErr   string
	}{
		{"unsafe permalink", "../evil", "", "unsafe permalink"},
		{"unsafe -name", "good", "../evil", "unsafe -name"},
		{"leading dash permalink", "-x", "", "unsafe permalink"},
		{"slash in -name", "good", "sub/clip", "unsafe -name"},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			client := &Client{}
			dir := t.TempDir()
			coub := Coub{Permalink: c.permalink}

			_, _, err := client.Download(context.Background(), coub, dir, c.name)
			if err == nil {
				t.Fatalf("Download(permalink=%q, name=%q) = nil error, want %q", c.permalink, c.name, c.wantErr)
			}
			if !strings.Contains(err.Error(), c.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, c.wantErr)
			}
			if entries, _ := os.ReadDir(dir); len(entries) != 0 {
				t.Errorf("Download wrote %d file(s), want none", len(entries))
			}
		})
	}
}

func TestDownloadNoVideoURL(t *testing.T) {
	client := &Client{}
	dir := t.TempDir()

	coub := Coub{Permalink: "test1"}

	_, skipped, err := client.Download(context.Background(), coub, dir, "")
	if err == nil {
		t.Fatal("Download returned nil error for a coub with no video URL")
	}
	if !strings.Contains(err.Error(), "no downloadable video") {
		t.Errorf("error = %q, want it to mention the missing video", err)
	}
	if skipped {
		t.Error("skipped = true, want false")
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("Download wrote %d file(s), want none", len(entries))
	}
}
