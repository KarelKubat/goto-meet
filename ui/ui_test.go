package ui

import (
	"goto-meet/cache"
	"goto-meet/item"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	for _, test := range []struct {
		notifier  string
		wantError string
	}{
		{
			notifier: "macos_osascript",
		},
		{
			notifier:  "nonsense",
			wantError: "no such notification type",
		},
	} {
		_, err := New(&Opts{
			Name: test.notifier,
		})
		switch {
		case err == nil && test.wantError != "":
			t.Errorf("New(for notifier %q) == _,nil, want error with %q", test.notifier, test.wantError)
		case err != nil && test.wantError == "":
			t.Errorf("New(for notifier %q) == _,%v, want nil error", test.notifier, err)
		case err != nil && test.wantError != "" && !strings.Contains(err.Error(), test.wantError):
			t.Errorf("New(for notifier %q) == _,%v, want error with %q", test.notifier, err, test.wantError)
		}
	}
}

func TestShouldSchedule(t *testing.T) {
	for _, test := range []struct {
		title       string
		startsIn    time.Duration
		joinLink    string
		wantOutcome bool
	}{
		{
			title:       "pass",
			startsIn:    time.Hour,
			joinLink:    "whatever",
			wantOutcome: true,
		},
		{
			title:       "in the past",
			startsIn:    -1 * time.Hour,
			joinLink:    "whatever",
			wantOutcome: false,
		},
		{
			title:       "no join link",
			startsIn:    time.Hour,
			wantOutcome: false,
		},
	} {
		n := &Notifier{
			opts: &Opts{
				StartsIn: time.Minute * 30, // consider anything that starts within half an hour
			},
			processed: cache.New(),
		}
		it := &item.Item{
			Title:    "whatever",
			JoinLink: test.joinLink,
			StartsIn: test.startsIn,
		}
		if outcome, _ := n.shouldSchedule(it); outcome != test.wantOutcome {
			t.Errorf("%v: shouldSchedule(%v) = %v, want %v", test.title, it, outcome, test.wantOutcome)
		}
	}
}
