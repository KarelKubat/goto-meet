package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KarelKubat/goto-meet/client"
	"github.com/KarelKubat/goto-meet/l"
	"github.com/KarelKubat/goto-meet/lib"
	"github.com/KarelKubat/goto-meet/lister"
	"github.com/KarelKubat/goto-meet/ui"
)

const (
	// Version of this package, increased upon releasing.
	version = "0.08"
)

var (
	// How to contact Google Calendar
	tokenFileFlag       = flag.String("token", "~/.goto-meet/token.json", "path to JSON configuration with access access_token etc., supports `~/` prefix")
	credentialsFileFlag = flag.String("credentials", "~/.goto-meet/credentials.json", "path to JSON configuration with client_id, project_id etc., supports '~/' prefix")
	clientTimeoutFlag   = flag.Duration("timeout", time.Second*30, "timeout when polling for new calendar entries, 0 to prevent timing out")

	// Calendar processing
	calendarsFlag      = flag.String("calendars", "primary", "comma-separated list of calendars to inspect, 'primary' is your default calendar")
	resultsPerPollFlag = flag.Int("results", 50, "max results to process per calendar poll")
	pollIntervalFlag   = flag.Duration("interval", time.Minute*10, "wait time between calendar polls")
	lookaheadFlag      = flag.Duration("look-ahead", time.Hour*1, "fetch calendar events that start before this duration")
	startsInFlag       = flag.Duration("starts-in", time.Minute, "how much in advance of a meeting should an alert be generated")

	// How to notify the user
	notificationTypeFlag = flag.String("notification", "macos_osascript", "type of notifications to generate")
	onscreenSecFlag      = flag.Int("onscreen-sec", 120, "number of seconds to keep a notification visible")
	browserFlag          = flag.String("browser", "", "browser to activate for calendar links, '' means default browser")

	// General
	loopsFlag    = flag.Int("loops", 0, "polling loops to execute before stopping, 0 means forever (mainly for debugging)")
	failuresFlag = flag.Int("failures", 10, "give up after # of consecutive polling errors")
	logFlag      = flag.String("log", "file://stdout", "logfile, see https://github.com/KarelKubat/smartlog")
	versionFlag  = flag.Bool("version", false, "show version and stop")
)

func main() {
	flag.Parse()
	if len(flag.Args()) > 0 {
		l.Fatalf("no positional arguments required, try `goto-meet --help`")
	}

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if err := l.SetOutput(*logFlag); err != nil {
		fmt.Fprintf(os.Stderr, "cannot set log destination: %v\n", err)
		os.Exit(1)
	}
	l.Infof("Welcome to goto-meet %v", version)

	tokenPath, err := lib.ExpandPath(*tokenFileFlag)
	if err != nil {
		l.Fatalf("%v", err)
	}

	l.Infof("path to token file: %v", tokenPath)
	credentialsPath, err := lib.ExpandPath(*credentialsFileFlag)
	if err != nil {
		l.Fatalf("%v", err)
	}
	l.Infof("path to credentials file: %v", credentialsPath)

	notifier, err := ui.New(&ui.Opts{
		Name:          *notificationTypeFlag,
		StartsIn:      *startsInFlag,
		VisibilitySec: *onscreenSecFlag,
		Browser:       *browserFlag,
	})
	if err != nil {
		l.Fatalf("%v", err)
	}

	ctx := context.Background()
	srv, err := client.New(ctx, &client.Opts{
		TokenFile:       tokenPath,
		CredentialsFile: credentialsPath,
		Timeout:         *clientTimeoutFlag,
	})
	if err != nil {
		l.Fatalf("cannot create client for the calendar service: %v", err)
	}
	lister, err := lister.New(ctx, &lister.Opts{
		Service:           srv,
		MaxResultsPerPoll: *resultsPerPollFlag,
		Calendars:         strings.Split(*calendarsFlag, ","),
		LookAhead:         *lookaheadFlag,
	})
	if err != nil {
		l.Fatalf("cannot create calendar lister: %v", err)
	}

	// Enter polling loop. Try to handle errors by only logging them until the max # of failures has been reached.
	nLoops := 0
	nFailures := 0
	for {
		// Quit after the indicated # of loops or when we've been failing all the time.
		nLoops++
		l.Infof("---------- Polling loop %v (%v consecutive errors) ----------", nLoops, nFailures)
		if *loopsFlag > 0 && nLoops > *loopsFlag {
			l.Warnf("exiting before loop %v", nLoops)
			break
		}

		// Get next entries and have the ui schedule alerts.
		if err := lister.Fetch(ctx); err != nil {
			nFailures++
			l.Warnf("failure %v: cannot fetch next calendar entries: %v", nFailures, err)
			if nFailures >= *failuresFlag {
				l.Fatalf("%v consecutive failures, giving up", nFailures)
			}
			time.Sleep(time.Second * 5)
			continue
		} else {
			nFailures = 0
		}

		for it := lister.First(); it != nil; it = lister.Next() {
			notifier.Schedule(it)
		}

		// Honor the polling interval, unless this is the first time around
		if nLoops > 1 {
			time.Sleep(*pollIntervalFlag)
		}
	}

	// Allow any notifications from the last loop to appear.
	time.Sleep(time.Second)
}
