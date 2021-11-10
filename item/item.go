// Package item wraps parsing and handling of a calendar item.
package item

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"goto-meet/lib"

	"google.golang.org/api/calendar/v3"
)

// Regexes for known URLs where we can find joining links.
var joinRegexes = []*regexp.Regexp{
	regexp.MustCompile(`.*href="(https://stream.meet.google.com/[^"]*).*"`),
	regexp.MustCompile(`.*href="(https://liveplayer.corp.google.com/[^"]*).*"`),
	regexp.MustCompile(`.*href="(http://go/watch[^"]*).*"`),
}

// Item is the receiver struct.
type Item struct {
	Event        *calendar.Event // item as returned by Google Calendar
	Title        string          // description of the event
	JoinLink     string          // extracted URL to join
	CalendarLink string          // extracted URL to see the calendar item
	Start        time.Time       // event start stamp
	StartsIn     time.Duration   // event start from now
}

// New creates an Item.
func New(event *calendar.Event) (*Item, error) {
	out := &Item{
		Event:        event,
		Title:        lib.Sanitize(event.Summary),
		CalendarLink: event.HtmlLink,
	}
	if ers := out.findStart(); ers != nil {
		return nil, ers
	}
	out.findJoinLink()

	return out, nil
}

// String returns a readable representation of an item.
func (i *Item) String() string {
	return fmt.Sprintf("%q (join:%v, calendar:%v, starts on %v, in:%v)",
		i.Title, i.JoinLink, i.CalendarLink, i.Start, i.StartsIn)
}

// findJoinLink is a helper to find a link to join a meeting in the calendar event.
func (i *Item) findJoinLink() {
	// Preferred is the hangout link, if absent, check the summary and description for known
	// URLs.
	if i.Event.HangoutLink != "" {
		i.JoinLink = i.Event.HangoutLink
		return
	}
	for _, s := range []string{
		i.Event.Summary,
		i.Event.Description,
	} {
		for _, re := range joinRegexes {
			match := re.FindStringSubmatch(s)
			if len(match) > 1 {
				i.JoinLink = match[1]
				return
			}
		}
	}
}

// findStart is a helper to extract the starting date/time of a calendar event.
func (i *Item) findStart() error {
	startCandidates := []string{}
	if i.Event.OriginalStartTime != nil {
		startCandidates = append(startCandidates,
			i.Event.OriginalStartTime.DateTime, i.Event.OriginalStartTime.Date)
	}
	if i.Event.Start != nil {
		startCandidates = append(startCandidates,
			i.Event.Start.DateTime, i.Event.Start.Date)
	}
	timeFound := false
	for _, start := range startCandidates {
		if start == "" {
			continue
		}
		timeFound = true
		if len(start) < 20 {
			start += "T00:00:00.000Z"
		}
		var err error
		i.Start, err = time.Parse(time.RFC3339, start)
		if err != nil {
			return fmt.Errorf("cannot parse timestamp %q: %v", start, err)
		}
	}
	if !timeFound {
		return errors.New("cannot find event start")
	}
	i.StartsIn = i.Start.Sub(time.Now())

	return nil
}
