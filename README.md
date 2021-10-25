# goto-meet

`goto-meet` is a command-line program that polls your Google Calendar and displays notifications of upcoming meetings
where you can join a video conference. The notifications have 3 buttons: one to join the video meeting, one
to see (or edit) the calendar entry, and one to ignore the notification.

Current limitations are the following:

- `goto-meet` expects that you have a Google Chrome browser.
- Notifications currently only work on MacOSX, because they use the `osascript` utility to render popups.

This version of `goto-meet` only scratches my own itch, but I may implement support for other browsers or
notifications as the need arises. Pull requests are of course always welcome. The wishlist, in abbreviated form:

- Add unit tests and make `goto-meet` a complete package
- The MacOSX notifications are a bit clunky. Is there a nicer way?
- If notifications allow this: can the browser be instructed to open on a given monitor?
- Implement notifications for other OSses.
- Add a method to prevent double invocations on non-MacOSX systems. Maybe `goto-meet` must become aware of its own
  PID file.

Questions / remarks? You can find me on karel@kubat.nl.

## Preparation

### Downloading and installing

- Download the goto-meet sources
- In the downloaded location, run `go mod init`
- Fetch required libraries:

```shell
go get -u google.golang.org/api/calendar/v3
go get -u golang.org/x/oauth2/google
```

- In the location where you put the `goto-meet` sources, run:

```shell
go build goto-meet.go              # build the binary goto-meet
sudo mv goto-meet /usr/local/bin/  # or use another appropriate location along your $PATH
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

The following steps are required just once. These instructions were written in October 2021 and may or may not still
be accurate as you read this text; Google may well have modified their website layout.

- Navigate to https://console.cloud.google.com/ and log in.
- Create a new project, and name it e.g. `CalendarAPI`. You may need to explicitly switch to this project if you
  have already other active projects.
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

These steps are required only once to allow `goto-meet` to consume the Google Calendar API.
Run `goto-meet`:

```shell
goto-meet            # fire it up
```

This will render a message that you should visit a location on accounts.google.com to fetch a code. Copy/paste the
shown link to your browser. Google will ask you whether you want to trust this `goto-meet` desktop app. Agree, and copy the generated code.

Back in the terminal, paste the code to the waiting `goto-meet` process and hit enter. The access code will be
saved as `~/.goto-meet/token.json`. `goto-meet` is now happily polling your calendar, but you can for now kill
the process and read on. Just hit ^C.

### First real try

For a testrun, try:

```shell
goto-meet --log='' --loops=3 --poll-interval=5s --starts-in=48h --look-ahead=72h
```

This will instruct `goto-meet` to do 3 polls, each 5 seconds apart. It will consider all events with a video
meeting within the next 3 days (72h) and will show a notification for each event that's starting within
the next 2 days (48h)
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

`goto-meet` tries to use "sane" defaults, but you can always use flags to modify its behavior. Try

```shell
goto-meet --help
```

to see what you can set. The following sections describe a few handy flags.

### Location of the config files

Use `--credentials-file` and `--token-file` to point `goto-meet` to different files than `credentials.json`
and `token.json` in the default location `~/.goto-meet/`. For example you could generate different configs
for different Google accounts and run several `goto-meet` processes to poll their calendars.

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

### Debugging

`goto-meet` writes its actions to a logfile, which is by default `/tmp/goto-meet.log`. Each time that `goto-meet`
starts, the log is overwritten. Use this flag to change the logfile location, or use `--log=''` to see
the log in the terminal.

## Automatic startup

The sources contain a file `nl.kubat.goto-meet.plist`. If you like `goto-meet` and want it running in the background:

- Copy the file to `~/Library/LaunchAgents/`
- Edit `~/Library/LaunchAgents/nl.kubat.goto-meet.plist` and make sure that the `ProgramArguments` array has all the
  flags that you want to set, and that the program is expected in the right location (`/usr/local/bin/` is assumed).
- Run `launchctl load ~/Library/LaunchAgents/nl.kubat.goto-meet.plist`.
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