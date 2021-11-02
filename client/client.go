// Package client wraps the functions instantiate a calendar client.
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Opts wraps the options to create a new client.
type Opts struct {
	TokenFile       string
	CredentialsFile string
	Timeout         time.Duration
}

// New returns an HTTP client, initialized either from an existing token file, or from the
// web (in which case a new token file is created).
func New(ctx context.Context, opts *Opts) (*calendar.Service, error) {
	// Sanity
	if opts.TokenFile == "" || opts.CredentialsFile == "" {
		return nil, errors.New("cannot instantiate client: token and credentials files must be given")
	}
	// Instantiate the configuration.
	b, err := ioutil.ReadFile(opts.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read %q", opts.CredentialsFile)
	}
	log.Printf("credentials file %q scannned", opts.CredentialsFile)
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		return nil, fmt.Errorf("cannot instantiate configuration: %v", err)
	}

	// Try to read the token file. If that fails, fetch a token from the web and create a new
	// token file.
	tok, haveTokenFile, err := tokenFromFile(opts.TokenFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read token file: %v", err)
	}
	if haveTokenFile {
		log.Printf("token file %q scanned", opts.TokenFile)
	} else {
		log.Printf("token file %q not found, fetching from web", opts.TokenFile)
		if tok, err = getTokenFromWeb(config); err != nil {
			return nil, fmt.Errorf("cannot read web token: %v", err)
		}
		if err = saveToken(opts.TokenFile, tok); err != nil {
			return nil, fmt.Errorf("cannot save web token to %q: %v", opts.TokenFile, err)
		}
	}

	// Instantiate a service connector.
	httpClient := config.Client(ctx, tok)
	httpClient.Timeout = opts.Timeout
	// log.Printf("HTTP client: %+v", *httpClient)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("cannot connecto to calendar service: %v", err)
	}
	return srv, nil
}

// tokenFromFile reads a token file and converts it to an oauth2.Token.
func tokenFromFile(file string) (*oauth2.Token, bool, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, false, nil
	}
	defer f.Close()
	tok := &oauth2.Token{}
	return tok, true, json.NewDecoder(f).Decode(tok)
}

// saveToken writes a web-obtained token to a file.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("cannot cache oauth token: %v", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(token); err != nil {
		return fmt.Errorf("cannot encode token to %q: %v", path, err)
	}
	return nil
}

// getTokenFromWeb fetches an online token. This code is run when a local tokenfile doesn't
// exist yet (or when it's faulty).
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("%s\n%s\n%s",
		"Go to the following link in your browser then type the authorization code:",
		authURL,
		"Enter the code followed by ENTER or hit ^C to abort: ")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("cannot read authorization code: %v", err)
	}
	if len(authCode) == 0 {
		return nil, errors.New("empty authorization code")
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %v", err)
	}
	return tok, nil
}
