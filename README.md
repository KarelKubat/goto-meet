# goto-meet

**`goto-meet` is a command-line program that polls your Google Calendar and displays notifications of upcoming meetings
containing a link to a video conference. The notifications have 3 buttons: one to join the video meeting, one
to see (or edit) the calendar entry, and one to ignore the notification.**

Like many similar utilities, `goto-meet` was born during the COVID19 lockdown period when meetings no longer
occurred in person and everything was via video chat. I wanted to have a straight forward notification system
that would pop up just prior to a video call, where I could click a *Join* button and be done with it --
 as opposed to firing up my calendar, searching for the event, and clicking a meeting button. `goto-meet` does
 exactly that. For any upcoming event, it will try to extract a video meet link and if found, will show a
 popup. The video meet link can be:

- The meeting link in the calendar event (called the `HangoutLink` in the Calendar API)
- Any link in the event's title or description that points to a "known" video service (see `item/item.go` in
  the sources).

Currently the limitation is that notifications currently only work on MacOSX, because they use the `osascript`
utility to render popups.

**Questions / remarks? You can find me on karel@kubat.nl.**

## Preparation

### Option 1: Downloading the sources, compiling and installing

You'll need a Go compiler and support to build your own binary.

- Download the `goto-meet` sources
- Review the sources if you want to check that `goto-meet` doesn't do anything malicious.
- In the downloaded location, run `go mod init`
- Fetch required libraries:

```shell
go get -u google.golang.org/api/calendar/v3
go get -u golang.org/x/oauth2/google
```

```shell
go build goto-meet.go              # build the binary goto-meet
sudo mv goto-meet /usr/local/bin/  # or use another appropriate location along your $PATH
```

### Option 2: Using a pre-built release

If you don't have a Go compiler and if you decide that a pre-built binary is okay, then just fetch a
release and unzip it. In it you will find:

- The `goto-meet` binary
- `nl.kubat.goto-meet.plist` which you may use on MacOSX to have `goto-meet` run as a background process
  (described later).

To install `goto-meet`, move it to an appropriate location along your `$PATH`:

```shell
sudo mv /where/ever/you/unzipped/it/goto-meet /usr/local/bin/ # or choose another location
```

### Default location for configs

`goto-meet` will expect its configuration to access the Google Calendar API in a directory `~/.goto-meet/`
(unless of course you use flags to point to different config files).
Create this location:

```shell
mkdir ~/.goto-meet       # create dir
chmod 700 ~/.goto-meet   # make it readable only by this user
```

### Enabling your Google Calendar API

The following steps are required just once. The purpose is to enable the Google Cloud API for your Google
account.

These instructions were written in October 2021 and may or may not still be accurate as you read this
text; Google may well have modified their website layout.

- Navigate to [console.cloud.google.com](https://console.cloud.google.com/) and log in.
- Create a new project, and name it e.g. `CalendarAPI`. If you already have Google cloud projects you can also
  put this API under the umbrella of an existing one, doesn't matter.
- On the card `Getting Started`, click `Explore and enable APIs`.
- Click the button `+ ENABLE APIS AND SERVICES`.
- In the row `Google Workspace` click `Google Calendar API`.
- Click the `Enable` button.
- On the left menu click `Credentials`.
- On the credentials screen click `+ CREATE CREDENTIALS` and choose `OAuth client ID`.
- On the screen to create an OAuth ID, click `CONFIGURE CONSENT SCREEN`.
- As user type, you may choose `Internal` so that the API will only work for users within your organization.
- Fill in the necessary data on the screen `Edit app registration`. This defines how the screen will look that asks
  for permission to access the calendar. At a minimum,
  - Set the `App name` to e.g. `goto-meet`,
  - Set the `User support email` to your email address,
  - Set the `Developer contact information` to your email address,
  - Click `SAVE AND CONTINUE`.
- You may modify the `Scopes` on the next screen to limit what the API is allowed to serve (we need just read-only
  access) or you can leave this screen as-is.
- Once you're done, click `BACK TO DASHBOARD`.
- Back on the dashboard, click for a second time `+ CREATE CREDENTIALS`, but now choose `OAuth client ID`.
- As application type choose `Desktop app`. Choose a name, `goto-meet` is an obvious candidate.
- Click `Create` and download the JSON file. It will be named something like `client_secret*.json` and will be in
  your download folder.
- Rename this file to `~/.goto-meet/credentials.json`:

```shell
  # Just an example, your downloads folder may be something different and make sure
  # to point to the downloaded `client_secret*` file.
  mv ~/Downloads/client_secret*json ~/.goto-meet/credentials.json`
```

### Authorizing goto-meet

These steps are required only once to allow `goto-meet` to consume the Google Calendar API that you enabled
in the step above.

Run `goto-meet`:

```shell
goto-meet --loops=1   # goto-meet won't poll repeatedly
```

This will show a message that you should visit a location on accounts.google.com to fetch a code.  Leave the
terminal as-is (waiting for input) and copy/paste the shown link to your browser. Google will ask you whether
you want to trust this `goto-meet` desktop app. Agree, and copy the generated code.

Back in the terminal, paste the code to the waiting `goto-meet` process and hit enter. The access code will be
saved as `~/.goto-meet/token.json`.

If this step fails for some reason, just remove `~/.goto-meet/token.json` and retry.

### First real try

For a testrun, try:

```shell
goto-meet --log='' --loops=3 --poll-interval=5s --starts-in=48h --look-ahead=72h
```

This will instruct `goto-meet` to do 3 polls, each 5 seconds apart. It will consider all events with a video
meeting within the next 3 days (72h) and will show a notification for each event that's starting within
the next 2 days (48h).
This of course assumes that you have a video meeting within that period. Adjust the flags `--look-ahead` and
`--starts-in` accordingly, until `goto-meet` finds something worth while.

NOTE: Durations are given as a number, followed by a prefix, such as `10s` or `20m` or `3h`. There is no suffix for
days, just use the number of days times 24, with `h` as the suffix. Different specifiers may also be
combined, as in `23h59m30s`, which is 30 seconds short of one full day.

The result should be an alert showing three buttons:

- *Join*, to open the video meeting link,
- *Calendar*, to open your calendar with the event,
- *Skip*, to dismiss the notification.

## Running it

`goto-meet` tries to use "sane" defaults, but you can always use flags to modify its behavior.
The following sections describe a few handy flags. To see an overview of all flags, try

```shell
goto-meet --help
```

### Calendar and polling

- `--calendar-id` tells `goto-meet` which calendar to poll. The default is `primary`, your main calendar, but you
  can choose a different one. If you want to poll several calendars, say your default and a calender `office`, then
  you can start two `goto-meet` processes, where one overrides the polled calendar using `--calendar-id=office`.
- `--starts-in` defines how long before an event a notification should be shown. The default is 1 minute.
- `--poll-interval` defines how long `goto-meet` waits between calendar polls. The default is 30 minutes; it's
  assumed that new calendar entries don't appear more frequently.
- `--look-ahead` defines how far ahead `goto-meet` looks when fetching new calendar entries. The default is 1 hour,
  meaning that each 30 minutes (the `--poll-interval`) the events for the next hour are fetched (the
`--look-ahead`).
- `--max-results-per-poll` limits the number of fetched entries during each poll. The default is 50, which assumes
  that you won't have more than 50 events within the next hour.

### UI

- `--onscreen-sec` defines how long a popup should remain visible. The default is 120.
- `--browser` identifies your favorite browser. The default is an empty string, which calls your default
  browser. This flag may be set to force accepting video calls in a different browser than your default one,
  e.g., on a different monitor.

### Location of the config files

Use `--credentials-file` and `--token-file` to point `goto-meet` to different files than `credentials.json`
and `token.json` in the default location `~/.goto-meet/`. For example you could generate different configs
for different Google accounts and run several `goto-meet` processes to poll their calendars.

### Debugging

`goto-meet` writes its actions to a logfile, which is by default `/tmp/goto-meet.log`. Each time that `goto-meet`
starts, the log is overwritten. Use this flag to change the logfile location, or use `--log=''` to see
the log in the terminal.

## Automatic startup

The sources contain a file `nl.kubat.goto-meet.plist`. If you like `goto-meet` and want it running in the background:

- Copy the file to `~/Library/LaunchAgents/`
- Edit `~/Library/LaunchAgents/nl.kubat.goto-meet.plist` and make sure that the `ProgramArguments` array has all the
  flags that you want to set, and that the program is expected in the right location (`/usr/local/bin/` is assumed).
- Run `launchctl start ~/Library/LaunchAgents/nl.kubat.goto-meet.plist`.
- Check that all worked:

```shell
ps ax | grep goto-meet   # one process must be running
cat /tmp/goto-meet.log   # the log must now exist
```

If you don't like MacOSX's `launchd` then you can just as easily fire up `goto-meet` by hand:

```shell
nohup goto-meet &  # fire up goto-meet as a background process
```

Or you can fire up `goto-meet` after each reboot by adding the following line to your crontab:

```shell
@reboot /usr/local/bin/goto-meet &
```

## TODOs

This version of `goto-meet` only scratches my own itch, but I may implement support for features as the need
arises. Pull requests are of course always welcome. The wishlist, in abbreviated form:

- Add unit tests and make `goto-meet` a complete package
- The MacOSX notifications are a bit clunky. Is there a nicer way?
- If notifications allow this: can the browser be instructed to open on a given monitor? `goto-meet` supports a
  work-around to force opening video meetings by another browser than your default one, but this still requires
  you to have two browsers open.
- Implement notifications for other OSses.
- Add a method to prevent double invocations on non-MacOSX systems. Maybe `goto-meet` must become aware of its own
  PID file.

## Version & Release Log (most recent last)

If you find a bug, please report it but also:

- State the version of this tool, you can find it using `goto-meet --version`
- Include the log, normally found as `/tmp/goto-meet.log`
- Clearly describe the problem.

Thanks!

```text
0.01 2021-10 Initial version-stamped release.
```
