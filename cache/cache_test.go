package cache

import (
	"fmt"
	"testing"
	"time"

	"github.com/KarelKubat/goto-meet/item"
)

func TestItemKey(t *testing.T) {
	now := time.Now()
	for _, test := range []struct {
		title        string
		joinLink     string
		calendarLink string
		start        time.Time
		wantKey      string
	}{
		{
			title:        "title",
			joinLink:     "joinLink",
			calendarLink: "calendarLink",
			start:        now,
			wantKey:      fmt.Sprintf("title::joinLink::calendarLink::%v", now.String()),
		},
	} {
		it := &item.Item{
			Title:        test.title,
			JoinLink:     test.joinLink,
			CalendarLink: test.calendarLink,
			Start:        now,
		}
		key := itemKey(it)
		if key != test.wantKey {
			t.Errorf("itemKey(%v) = %v, want %v", it, key, test.wantKey)
		}
	}
}

func TestLookup(t *testing.T) {
	now := time.Now()
	c := New()
	items := []struct {
		title        string
		joinLink     string
		calendarLink string
	}{
		{
			title:        "1",
			joinLink:     "2",
			calendarLink: "3",
		},
		{
			title:        "4",
			joinLink:     "5",
			calendarLink: "6",
		},
	}

	// Insert items, they must be added
	for _, test := range items {
		it := &item.Item{
			Title:        test.title,
			JoinLink:     test.joinLink,
			CalendarLink: test.calendarLink,
			Start:        now,
		}
		if c.Lookup(it) {
			t.Errorf("Lookup(%v) = true", it)
		}
	}
	// Do it again, they must already be present
	for _, test := range items {
		it := &item.Item{
			Title:        test.title,
			JoinLink:     test.joinLink,
			CalendarLink: test.calendarLink,
			Start:        now,
		}
		if !c.Lookup(it) {
			t.Errorf("Lookup(%v) = false", it)
		}
	}
}

func TestWeed(t *testing.T) {
	before := time.Now().Add(time.Hour * -1)
	after := time.Now().Add(time.Hour)
	c := New()

	items := []struct {
		title        string
		joinLink     string
		calendarLink string
		start        time.Time
		wantFound    bool // want this to be found after weeding?
	}{
		{
			title:        "1",
			joinLink:     "2",
			calendarLink: "3",
			start:        after,
			wantFound:    true,
		},
		{
			title:        "4",
			joinLink:     "5",
			calendarLink: "6",
			start:        before,
			wantFound:    false,
		},
	}

	// Add items, all must be flagged as "added"
	for _, test := range items {
		it := &item.Item{
			Title:        test.title,
			JoinLink:     test.joinLink,
			CalendarLink: test.calendarLink,
			Start:        test.start,
		}
		if c.Lookup(it) {
			t.Errorf("Lookup(%v) = false", it)
		}
	}
	c.Weed()
	for _, test := range items {
		it := &item.Item{
			Title:        test.title,
			JoinLink:     test.joinLink,
			CalendarLink: test.calendarLink,
			Start:        test.start,
		}
		if found := c.Lookup(it); found != test.wantFound {
			t.Errorf("Lookup(%v) = %v, want %v", it, found, test.wantFound)
		}
	}
}
