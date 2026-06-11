package coub

import "testing"

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
