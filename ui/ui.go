// Package ui is responsible for the user interface.
package ui

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"
	"time"

	"github.com/KarelKubat/goto-meet/cache"
	"github.com/KarelKubat/goto-meet/item"
	"github.com/KarelKubat/goto-meet/l"
)

const (
	// Interval between clock skew checks
	heartbeatInterval = time.Second * 10
)

// The name, system command and corresponding template for a notification.
type notificationSettings struct {
	name string
	args []string
	tpl  *template.Template
}

var (
	// UI notifications. Use {{.Browser}} and so on to fill in. The `args`
	// program will be called with the expanded `tpl` on stdin.
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
	Name          string        // Name of this notifier
	StartsIn      time.Duration // Duration before the event to render the UI
	VisibilitySec int           // How long the UI should stay visible
	Browser       string        // Browser to call upon "join"
}

// Notifier wraps the applicable notification configuration.
type Notifier struct {
	opts      *Opts                 // Name, lead time etc. to show an alert before a meeting starts
	config    *notificationSettings // One of the notificationConfigs
	processed *cache.Cache          // Has an event been processed yet?
}

// New creates a Notifier.
func New(opts *Opts) (*Notifier, error) {
	availableNotificationTypes := []string{}
	for _, config := range notificationConfig {
		if config.name == opts.Name {
			// Matched the requested UI notifier.
			out := &Notifier{
				config:    config,
				opts:      opts,
				processed: cache.New(),
			}
			// Start the heartbeat to remove cached entries when a clock skew is detected.
			go func() {
				for {
					start := time.Now()
					time.Sleep(heartbeatInterval)
					// Unconsciousness for more than 1 second will be detected.
					if time.Now().After(start.Add(heartbeatInterval + time.Second)) {
						l.Infof("time skew detected")
						out.processed.Clear()
					}
				}
			}()
			l.Infof("notifier %q created to alert %v before event start", opts.Name, opts.StartsIn)
			return out, nil
		}
		availableNotificationTypes = append(availableNotificationTypes, config.name)
	}

	// No handler found for the requested UI notifier.
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
	n.processed.Weed()
	toSchedule, waitTime := n.shouldSchedule(it)
	if !toSchedule {
		return
	}

	go func(it *item.Item) {
		l.Infof("notification in %v for event %v", waitTime, it)
		time.Sleep(waitTime)

		// We've woken up and it's time to show a notification. In the meantime the laptop might have
		// gone to sleep and woken up way past the the starttime of the event - in which case we just return.
		//
		// Fortunately there's a heartbeat that detects clock skew and clears the cache, so that events are
		// re-scheduled into another go-routine. So this event notifier may fail, there will be a backup.
		if time.Now().After(it.Start.Add(time.Second)) {
			l.Infof("skipping notifiying for %v, it's too much in the past", it)
			return
		}

		t := &temp{
			Title:         it.Title,
			VisibilitySec: n.opts.VisibilitySec,
			Browser:       n.opts.Browser,
			JoinLink:      it.JoinLink,
			CalendarLink:  it.CalendarLink,
		}
		buf := new(bytes.Buffer)
		if err := n.config.tpl.Execute(buf, t); err != nil {
			l.Warnf("cannot execute template: %v", err)
			return
		}
		l.Infof("template: %v", buf.String())
		cmd := exec.Command(n.config.args[0], n.config.args[1:]...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			l.Warnf("cannot create pipe to notifier: %v", err)
			return
		}
		go func() {
			defer stdin.Close()
			_, err := stdin.Write(buf.Bytes())
			if err != nil {
				l.Warnf("failed to write to notifier: %v", err)
			}
		}()
		out, err := cmd.CombinedOutput()
		if err != nil {
			l.Warnf("notifier failed, output: %v, error: %v", string(out), err)
			return
		}
	}(it)
}

// shouldSchedule is a helper to determine whether an item is worthy of scheduling.
func (n *Notifier) shouldSchedule(it *item.Item) (bool, time.Duration) {
	switch {
	case it.StartsIn < 0:
		l.Infof("%q starts in the past, not worthy scheduling; start: %v", it.Title, it.Start)
		return false, 0
	case it.JoinLink == "":
		l.Infof("%v has no join link, not worthy scheduling; entry: %v", it, it.Event)
		return false, 0
	case n.processed.Lookup(it):
		l.Infof("%v already processed, not worthy (re)scheduling", it)
		return false, 0
	default:
		return true, it.StartsIn - n.opts.StartsIn
	}
}
