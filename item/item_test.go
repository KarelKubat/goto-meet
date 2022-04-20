package item

import (
	"testing"

	"google.golang.org/api/calendar/v3"
)

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
			summary:      `<a href="http://go/cp-sre-townhall-livestream" id="ow3404" __is_owner="true">Livestream</a>`,
			wantJoinLink: "http://go/cp-sre-townhall-livestream",
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
