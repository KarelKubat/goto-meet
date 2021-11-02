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

func TestSanitize(t *testing.T) {
	for _, test := range []struct {
		in      string
		wantOut string
	}{
		{
			// no sanitizing necessary
			in:      "whatever",
			wantOut: "whatever",
		},
		{
			// one quote
			in:      "what'ever",
			wantOut: "whatever",
		},
		{
			// multiple quotes
			in:      "w'h'a't'e'v'e'r",
			wantOut: "whatever",
		},
		{
			// weird chars are left as-is
			in:      `[]{}"/_.`,
			wantOut: `[]{}"/_.`,
		},
	} {
		if out := Sanitize(test.in); out != test.wantOut {
			t.Errorf("Sanitize(%q) = %q, want %q", test.in, out, test.wantOut)
		}
	}
}
