package ui

import (
	"strings"
	"testing"
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
