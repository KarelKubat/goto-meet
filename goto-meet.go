package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"goto-meet/client"
	"goto-meet/lister"
	"goto-meet/ui"
)

const (
	// Version of this package, increased upon releasing.
	version = "0.01"
)

var (
	// How to contact Google Calendar
	tokenFileFlag       = flag.String("token-file", "~/.goto-meet/token.json", "path to JSON configuration with access access_token etc., supports `~/` prefix")
	credentialsFileFlag = flag.String("credentials-file", "~/.goto-meet/credentials.json", "path to JSON configuration with client_id, project_id etc., supports '~/' prefix")
	clientTimeoutFlag   = flag.Duration("client-timeout", time.Second*30, "timeout when polling for new calendar entries, 0 to prevent timing out")

	// Calendar processing
	calendarIDFlag     = flag.String("calendar-id", "primary", "calendar to inspect, 'primary' is your default calendar")
	resultsPerPollFlag = flag.Int("max-results-per-poll", 50, "max results to process per calendar poll")
	pollIntervalFlag   = flag.Duration("poll-interval", time.Minute*10, "wait time between calendar polls")
	lookaheadFlag      = flag.Duration("look-ahead", time.Hour*1, "fetch calendar events that start before this duration")
	startsInFlag       = flag.Duration("starts-in", time.Minute, "how much in advance of a meeting should an alert be generated")

	// How to notify the user
	notificationTypeFlag = flag.String("notification-type", "macos_osascript", "type of notifications to generate")
	onscreenSecFlag      = flag.Int("onscreen-sec", 120, "number of seconds to keep a notification visible")
	browserFlag          = flag.String("browser", "", "browser to activate for calendar links, '' means default browser")

	// General
	loopsFlag    = flag.Int("loops", 0, "polling loops to execute before stopping, 0 means forever (mainly for debugging)")
	failuresFlag = flag.Int("failures", 10, "give up after # of consecutive polling errors")
	logfileFlag  = flag.String("log", "/tmp/goto-meet.log", "log to (over)write, use '' for stdout")
	versionFlag  = flag.Bool("version", false, "show version and stop")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *logfileFlag != "" {
		logFile, err := os.Create(*logfileFlag)
		if err != nil {
			log.Fatalf("cannot create log file: %v", err)
		}
		log.SetOutput(logFile)
	}

	tokenPath, err := expandPath(*tokenFileFlag)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("path to token file: %v", tokenPath)
	credentialsPath, err := expandPath(*credentialsFileFlag)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("path to credentials file: %v", credentialsPath)

	notifier, err := ui.New(&ui.Opts{
		Name:          *notificationTypeFlag,
		StartsIn:      *startsInFlag,
		VisibilitySec: *onscreenSecFlag,
		Browser:       *browserFlag,
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	srv, err := client.New(ctx, &client.Opts{
		TokenFile:       tokenPath,
		CredentialsFile: credentialsPath,
		Timeout:         *clientTimeoutFlag,
	})
	if err != nil {
		log.Fatalf("cannot create client for the calendar service: %v", err)
	}
	lister, err := lister.New(&lister.Opts{
		Service:           srv,
		MaxResultsPerPoll: *resultsPerPollFlag,
		CalendarID:        *calendarIDFlag,
		LookAhead:         *lookaheadFlag,
	})
	if err != nil {
		log.Fatalf("cannot create calendar lister: %v", err)
	}

	// Enter polling loop. Try to handle errors by only logging them until the max # of failures has been reached.
	nLoops := 0
	nFailures := 0
	for {
		// Quit after the indicated # of loops or when we've been failing all the time.
		nLoops++
		if *loopsFlag > 0 && nLoops > *loopsFlag {
			log.Printf("exiting before loop %v", nLoops)
			break
		}
		log.Printf("polling loop %v, consecutive polling errors: %v", nLoops, nFailures)

		// Get next entries and have the ui schedule alerts.
		if err := lister.Fetch(ctx); err != nil {
			nFailures++
			log.Printf("failure %v: cannot fetch next calendar entries: %v", nFailures, err)
			if nFailures >= *failuresFlag {
				log.Fatalf("%v consecutive failures, giving up", nFailures)
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

// expandPath is a helper to expand typical Unix shortands in paths.
func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~/") {
		usr, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("cannot find user's homedir: %v", err)
		}
		return filepath.Join(usr.HomeDir, p[2:]), nil
	}
	return p, nil
}
