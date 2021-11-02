package main

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
		out, err := expandPath(test.in)
		if err != nil {
			t.Fatalf("expandPath(%q) = _,%v, require nil error", test.in, nil)
		}
		re := regexp.MustCompile(test.wantOut)
		if !re.MatchString(out) {
			t.Errorf("expandPath(%q) = %q,_, want match with %q", test.in, out, test.wantOut)
		}
	}
}
