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
// This allows token refresh without requiring the client-secret.json file
type TokenFile struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	// Client credentials for self-sufficient token refresh
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

func NewTokenFile(ClientID, ClientSecret string) *TokenFile {
	return &TokenFile{ClientID: ClientID, ClientSecret: ClientSecret}
}

func (t *TokenFile) Refresh(token *oauth2.Token) *TokenFile {
	t.AccessToken = token.AccessToken
	t.TokenType = token.TokenType
	t.RefreshToken = token.RefreshToken
	t.Expiry = token.Expiry
	return t
}

// Retrieves a token from a local file.
// Supports both enhanced TokenFile format and standard oauth2.Token format
func tokenFromFile(file string) (*oauth2.Token, *TokenFile, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// Check if it's a directory
	info, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	if info.IsDir() {
		return nil, nil, fmt.Errorf("token path '%s' is a directory, expected a file", file)
	}

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
func saveToken(path string, token *TokenFile) {
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

	json.NewEncoder(f).Encode(token)
}
