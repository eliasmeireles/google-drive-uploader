package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/eliasferreira/google-drive-uploader/internal/config"
)

// Authenticator handles the OAuth2 authentication process
type Authenticator struct {
	Config config.Config
}

// NewAuthenticator creates a new Authenticator
func NewAuthenticator(config config.Config) *Authenticator {
	return &Authenticator{
		Config: config,
	}
}

// GetClient retrieves a token, saves the token, then returns the generated client.
func (a *Authenticator) GetClient(ctx context.Context, scope ...string) (*http.Client, error) {
	// If no scope is provided, default to DriveScope (full access)
	if len(scope) == 0 {
		scope = []string{"https://www.googleapis.com/auth/drive"}
	}

	// Check if we have a valid token first
	tok, tokenFile, err := tokenFromFile(a.Config.TokenPath)
	if err != nil && !a.Config.TokenGen {
		return nil, err
	}

	if !a.Config.TokenGen {
		// We have a valid token, check if it has embedded client credentials
		var config *oauth2.Config
		// If token has embedded credentials, use them (self-sufficient token)
		if tokenFile.ClientID != "" && tokenFile.ClientSecret != "" {
			config = &oauth2.Config{
				ClientID:     tokenFile.ClientID,
				ClientSecret: tokenFile.ClientSecret,
				Scopes:       scope,
				Endpoint:     google.Endpoint,
			}
			client := a.getClient(ctx, config, tok)
			return client, nil
		}
	}

	// No valid token, we need the client config to authenticate
	if a.Config.ClientSecret == "" {
		return nil, fmt.Errorf("no valid token found and --client-secret not provided")
	}

	b, err := os.ReadFile(a.Config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v\nPlease provide a valid client secret file by using --client-secret", err)
	}

	config, err := google.ConfigFromJSON(b, scope...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client := a.getClient(ctx, config, nil)
	return client, nil
}

// getClient retrieves a token, saves the token, then returns the generated client.
func (a *Authenticator) getClient(ctx context.Context, config *oauth2.Config, existingToken *oauth2.Token) *http.Client {
	var tok *oauth2.Token

	if existingToken != nil {
		tok = existingToken
	} else {
		t, tokenData, err := tokenFromFile(a.Config.TokenPath)
		if err != nil {
			fmt.Printf("No token found at %s, starting authorization flow...\n", a.Config.TokenPath)
			tok = getTokenFromWeb(config)
			tokenData = NewTokenFile(config.ClientID, config.ClientSecret)
			saveToken(a.Config.TokenPath, tokenData.Refresh(tok))
		} else {
			tok = t
		}
	}

	// Create a token source that will automatically refresh and save the token
	ts := config.TokenSource(ctx, tok)
	wrappedTs := &savingTokenSource{
		source: ts,
		path:   a.Config.TokenPath,
		config: config,
	}

	// Check if token is expired or will expire soon and refresh immediately if so
	// This ensures we have a valid token before we start any operation
	initialTok, err := wrappedTs.Token()
	if err != nil {
		fmt.Printf("Failed to refresh token: %v. Requesting new authorization...\n", err)
		tok = getTokenFromWeb(config)
		tokenData := NewTokenFile(config.ClientID, config.ClientSecret)
		saveToken(a.Config.TokenPath, tokenData.Refresh(tok))
		// Update the wrapped source with the new token
		ts = config.TokenSource(ctx, tok)
		wrappedTs.source = ts
		initialTok = tok
	}

	return oauth2.NewClient(ctx, oauth2.ReuseTokenSource(initialTok, wrappedTs))
}
