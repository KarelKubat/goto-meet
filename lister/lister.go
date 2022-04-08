// Package lister encapsulates fetching items from the Google Calendar.
package lister

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/KarelKubat/goto-meet/item"
	"github.com/KarelKubat/goto-meet/l"

	"google.golang.org/api/calendar/v3"
)

// Opts wraps paramenters when creating a lister.
type Opts struct {
	Service           *calendar.Service
	MaxResultsPerPoll int
	Calendars         []string
	LookAhead         time.Duration
}

// List represents fetched items that we can iterate on.
type List struct {
	Items []*item.Item
	index int
}

// Lister is the receiver.
type Lister struct {
	opts *Opts
	list *List
}

// New creates a Lister.
func New(ctx context.Context, opts *Opts) (*Lister, error) {
	// Sanity checks for the options
	if opts.Service == nil {
		return nil, errors.New("cannot instantiate a lister with a nil service")
	}
	if opts.MaxResultsPerPoll > 250 {
		return nil, errors.New("the maximum number of entries to fetch is limited to 250")
	}
	if len(opts.Calendars) == 0 {
		return nil, errors.New("there must be at least one calendar to check")
	}

	// Verify that the user's calendars exist.
	cals, err := opts.Service.CalendarList.
		List().
		Context(ctx).
		ShowDeleted(false).
		Do()
	if err != nil {
		return nil, errors.New("cannot list user's calendars")
	}
	availableMap := map[string]struct{}{
		"primary": {}, // "primary" always exists
	}
	availableNames := []string{}
	for _, it := range cals.Items {
		availableMap[it.Id] = struct{}{}
		availableNames = append(availableNames, it.Id)
	}
	for _, cal := range opts.Calendars {
		if _, ok := availableMap[cal]; !ok {
			return nil, fmt.Errorf("no such calendar %q, available: %v", cal, availableNames)
		}
	}

	l.Infof("calendar lister will look ahead %v and fetch max %v entries each run", opts.LookAhead, opts.MaxResultsPerPoll)
	return &Lister{
		opts: opts,
	}, nil
}

// Fetch polls for pending items and populates the list to process.
func (lis *Lister) Fetch(ctx context.Context) error {
	timeMin := time.Now().Format(time.RFC3339)
	timeMax := time.Now().Add(lis.opts.LookAhead).Format(time.RFC3339)

	lis.list = &List{}
	for _, calendar := range lis.opts.Calendars {
		events, err := lis.opts.Service.Events.
			List(calendar).
			MaxResults(int64(lis.opts.MaxResultsPerPoll)).
			ShowDeleted(false).
			Context(ctx).
			SingleEvents(true).
			TimeMin(timeMin).
			TimeMax(timeMax).
			OrderBy("startTime").
			Do()
		// TODO: skip not fully accepted entries where the user is a "maybe"
		if err != nil {
			return fmt.Errorf("unable to retrieve next %v events for calendar %q: %v", lis.opts.MaxResultsPerPoll, calendar, err)
		}
		for _, it := range events.Items {
			i, err := item.New(it)
			if err != nil {
				return fmt.Errorf("cannot initialize calendar item: %v", err)
			}
			lis.list.Items = append(lis.list.Items, i)
		}
		l.Infof("calendar %v: %v upcoming events", calendar, len(lis.list.Items))
	}

	return nil
}

// First returns the first fetched item, or nil.
func (l *Lister) First() *item.Item {
	l.list.index = 0
	if len(l.list.Items) < 1 {
		return nil
	}
	return l.list.Items[0]
}

// Next returns the next fetched item, or nil.
func (l *Lister) Next() *item.Item {
	l.list.index++
	if l.list.index >= len(l.list.Items) {
		return nil
	}
	return l.list.Items[l.list.index]
}
