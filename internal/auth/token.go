package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

// TokenFile represents the enhanced token format with embedded client credentials
// This allows token refresh without requiring the api-key.json file
type TokenFile struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	// Client credentials for self-sufficient token refresh
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

// Retrieves a token from a local file.
// Supports both enhanced TokenFile format and standard oauth2.Token format
func tokenFromFile(file string) (*oauth2.Token, *TokenFile, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// Try to read as TokenFile first (enhanced format)
	tokenFile := &TokenFile{}
	err = json.NewDecoder(f).Decode(tokenFile)
	if err != nil {
		return nil, nil, err
	}

	// Convert to oauth2.Token
	tok := &oauth2.Token{
		AccessToken:  tokenFile.AccessToken,
		TokenType:    tokenFile.TokenType,
		RefreshToken: tokenFile.RefreshToken,
		Expiry:       tokenFile.Expiry,
	}

	return tok, tokenFile, nil
}

// Saves a token to a file path with optional OAuth config for enhanced format
func saveToken(path string, token *oauth2.Token, config *oauth2.Config) {
	fmt.Printf("Saving credential file to: %s\n", path)
	// Ensure directory exists
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Unable to create directory for token: %v\n", err)
			return
		}
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Unable to cache oauth token: %v", err)
		return
	}
	defer f.Close()

	// Create enhanced token file with client credentials if config is provided
	tokenFile := TokenFile{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	// Extract client credentials from config if available
	if config != nil {
		tokenFile.ClientID = config.ClientID
		tokenFile.ClientSecret = config.ClientSecret
	}

	json.NewEncoder(f).Encode(tokenFile)
}
