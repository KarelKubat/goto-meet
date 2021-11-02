package client

import (
	"context"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestNew(t *testing.T) {
	for _, test := range []struct {
		tokenFile       string
		credentialsFile string
		wantError       string
	}{
		{
			// Both token and credentials filenames must be given
			wantError: "token and credentials files must be given",
		},
		{
			tokenFile: "whatever",
			wantError: "token and credentials files must be given",
		},
		{
			credentialsFile: "whatever",
			wantError:       "token and credentials files must be given",
		},
		{
			// credentials file must be readable
			tokenFile:       "whatever",
			credentialsFile: "/non/existing",
			wantError:       "cannot read",
		},
	} {
		_, err := New(context.Background(), &Opts{
			TokenFile:       test.tokenFile,
			CredentialsFile: test.credentialsFile,
		})
		if err == nil {
			t.Fatalf("New(%+v) = _,nil, want error", test)
		}
		if !strings.Contains(err.Error(), test.wantError) {
			t.Errorf("New(%+v) = _,%v, want error with %q", test, err, test.wantError)
		}
	}
}

func TestFileHandling(t *testing.T) {
	now := time.Now()
	dummy := "/tmp/dummy.json"
	tok := &oauth2.Token{
		AccessToken:  "accesstoken",
		TokenType:    "tokentype",
		RefreshToken: "refreshtoken",
		Expiry:       now,
	}

	if err := saveToken(dummy, tok); err != nil {
		t.Fatalf("saveToken(%q, _) = %v, require nil error", dummy, err)
	}
	newTok, got, err := tokenFromFile(dummy)
	if err != nil {
		t.Fatalf("tokenFromFile(%q) = _,_,%v, require nil error", dummy, err)
	}
	if !got {
		t.Fatalf("tokenFromFile(%q) = _,false,_, want true", dummy)
	}
	if newTok.AccessToken != tok.AccessToken {
		t.Errorf("access token mismatch: got %q, want %q", newTok.AccessToken, tok.AccessToken)
	}
	if newTok.TokenType != tok.TokenType {
		t.Errorf("token type mismatch: got %q, want %q", newTok.TokenType, tok.TokenType)
	}
	if newTok.RefreshToken != tok.RefreshToken {
		t.Errorf("refresh token mismatch: got %q, want %q", newTok.RefreshToken, tok.RefreshToken)
	}
	if !newTok.Expiry.Equal(tok.Expiry) {
		t.Errorf("token expiry mismatch: got %v, want %q=v", newTok.Expiry, tok.Expiry)
	}
}
