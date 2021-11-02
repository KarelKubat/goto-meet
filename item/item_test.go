package item

import (
	"testing"

	"google.golang.org/api/calendar/v3"
)

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
		if out := sanitize(test.in); out != test.wantOut {
			t.Errorf("sanitize(%q) = %q, want %q", test.in, out, test.wantOut)
		}
	}
}

func TestFindJoinLink(t *testing.T) {
	for _, test := range []struct {
		hangoutLink  string
		summary      string
		description  string
		wantJoinLink string
	}{
		{
			// Hangout links are taken as-is and take overall precedence
			hangoutLink:  "abc",
			summary:      `<a href="https://stream.meet.google.com/what/ever">`,
			description:  `<a href="https://liveplayer.corp.google.com/what/ever">`,
			wantJoinLink: "abc",
		},
		{
			// Summaries are examined for patterns
			summary:      `<a href="https://stream.meet.google.com/what/ever">`,
			wantJoinLink: "https://stream.meet.google.com/what/ever",
		},
		{
			summary:      `<a href="https://liveplayer.corp.google.com/what/ever">`,
			wantJoinLink: "https://liveplayer.corp.google.com/what/ever",
		},
		{
			// Descriptions are examined for patterns
			description:  `<a href="https://stream.meet.google.com/what/ever">`,
			wantJoinLink: "https://stream.meet.google.com/what/ever",
		},
		{
			description:  `<a href="https://liveplayer.corp.google.com/what/ever">`,
			wantJoinLink: "https://liveplayer.corp.google.com/what/ever",
		},
		{
			// Summaries take precendence over descriptions
			summary:      `<a href="https://stream.meet.google.com/what/ever">`,
			description:  `<a href="https://liveplayer.corp.google.com/what/ever">`,
			wantJoinLink: "https://stream.meet.google.com/what/ever",
		},
	} {
		it := &Item{
			Event: &calendar.Event{
				HangoutLink: test.hangoutLink,
				Summary:     test.summary,
				Description: test.description,
			},
		}
		it.findJoinLink()
		if it.JoinLink != test.wantJoinLink {
			t.Errorf("findJoinLink with event %+v = %q, want %q", it.Event, it.JoinLink, test.wantJoinLink)
		}
	}
}

// TODO: add findStart() test
