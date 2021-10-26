// Package ui is responsible for the user interface.
package ui

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"text/template"
	"time"

	"goto-meet/item"
)

// The name, system command and corresponding template for a notification.
type notificationSettings struct {
	name string
	args []string
	tpl  *template.Template
}

var (
	notificationConfig = []*notificationSettings{
		{
			name: "macos_osascript",
			args: []string{"osascript"},
			tpl: template.Must(template.New("macos_osascript").Parse(`
display dialog ("{{.Title}}") buttons {"Join", "Calendar", "Skip"} giving up after {{.VisibilitySec}}
if button returned of result = "Join" then
  {{if .Browser }}
  tell application "{{.Browser}}"
    activate
    open location "{{.JoinLink}}"
  end tell
  {{ else }}
  open location "{{.JoinLink}}"
  {{ end }}
else if button returned of result = "Calendar" then
  {{if .Browser }}
  tell application "{{.Browser}}"
    activate
    open location "{{.CalendarLink}}"
  end tell
  {{ else }}
  open location "{{.CalendarLink}}"
  {{ end }}
end if
`)),
		},
	}
)

// Opts wraps the options to create a notifier.
type Opts struct {
	Name          string
	StartsIn      time.Duration
	VisibilitySec int
	Browser       string
}

// Notifier wraps the applicable notification configuration.
type Notifier struct {
	opts   *Opts                 // name and lead time to show an alert before a meeting starts
	config *notificationSettings // one of the notificationConfigs
	cache  map[string]struct{}   // has an alert been shown yet?
}

// New creates a Notifier.
func New(opts *Opts) (*Notifier, error) {
	availableNotificationTypes := []string{}
	for _, config := range notificationConfig {
		if config.name == opts.Name {
			log.Printf("notifier %q created to alert %v before event start", opts.Name, opts.StartsIn)
			return &Notifier{
				config: config,
				opts:   opts,
				cache:  map[string]struct{}{},
			}, nil
		}
		availableNotificationTypes = append(availableNotificationTypes, config.name)
	}
	return nil, fmt.Errorf("no such notification type %q, choose one of %v", opts.Name, availableNotificationTypes)
}

// temp is used in template expansion.
type temp struct {
	Title         string // event title
	VisibilitySec int    // # secs on screen
	Browser       string // browser to fire up
	JoinLink      string // link to join the meet
	CalendarLink  string // link to see the event on the calendar
}

// Show notifies the user of an upcoming event.
func (n *Notifier) Schedule(it *item.Item) {
	key := fmt.Sprintf("%v::%v::%v", it.Title, it.JoinLink, it.CalendarLink)
	if _, ok := n.cache[key]; ok {
		log.Printf("notification %q already processed", key)
		return
	}
	n.cache[key] = struct{}{}

	go func() {
		waitTime := it.StartsIn - n.opts.StartsIn
		log.Printf("notification in %v for event %q, starts on %v (in %v, joinlink %q, calendarlink %q)",
			waitTime, it.Title, it.Start, it.StartsIn, it.JoinLink, it.CalendarLink)

		// Don't act on calendar items that lack a joinlink.
		if it.JoinLink == "" {
			log.Printf("no meeting link, skipped")
			return
		}

		// Wait until the calendar event is about to start.
		time.Sleep(waitTime)

		// Generate and render a notification.
		t := &temp{
			Title:         it.Title,
			VisibilitySec: n.opts.VisibilitySec,
			Browser:       n.opts.Browser,
			JoinLink:      it.JoinLink,
			CalendarLink:  it.CalendarLink,
		}
		buf := new(bytes.Buffer)
		if err := n.config.tpl.Execute(buf, t); err != nil {
			log.Printf("WARNING: cannot execute template: %v", err)
			return
		}
		log.Printf("template: %v", buf.String())
		cmd := exec.Command(n.config.args[0], n.config.args[1:]...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Printf("WARNING: cannot create pipe to notifier: %v", err)
			return
		}
		go func() {
			defer stdin.Close()
			_, err := stdin.Write(buf.Bytes())
			if err != nil {
				log.Printf("WARNING: failed to write to notifier: %v", err)
			}
		}()
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("WARNING: notifier failed, output: %v, error: %v", string(out), err)
			return
		}
	}()
}
