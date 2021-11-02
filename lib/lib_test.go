package lib

import (
	"regexp"
	"testing"
)

func TestExpandPath(t *testing.T) {
	for _, test := range []struct {
		in      string
		wantOut string
	}{
		{
			// No leading ~/, returned as-is
			in:      "/what/ever",
			wantOut: "^/what/ever$",
		},
		{
			// Leading ~/, returned as path (first rune /)
			in:      "~/what/ever",
			wantOut: "^/.*ever$",
		},
	} {
		out, err := ExpandPath(test.in)
		if err != nil {
			t.Fatalf("ExpandPath(%q) = _,%v, require nil error", test.in, nil)
		}
		re := regexp.MustCompile(test.wantOut)
		if !re.MatchString(out) {
			t.Errorf("ExpandPath(%q) = %q,_, want match with %q", test.in, out, test.wantOut)
		}
	}
}
