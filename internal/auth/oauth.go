package auth

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"
)

// Request a token from the web, then returns the retrieved token.
// Uses local callback server with automatic browser opening, falls back to manual flow if needed.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Try to use callback server approach
	port, err := findAvailablePort()
	if err != nil {
		fmt.Printf("Warning: Could not find available port: %v\n", err)
		fmt.Println("Falling back to manual authorization flow...")
		return getTokenFromWebManual(config)
	}

	// Update config to use local callback
	originalRedirectURL := ""
	if len(config.RedirectURL) > 0 {
		originalRedirectURL = config.RedirectURL
	}
	config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	// Start callback server
	codeChan, errChan, server := startCallbackServer(port)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// Give server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Generate auth URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Try to open browser
	fmt.Printf("Opening browser for authorization...\n")
	fmt.Printf("If the browser doesn't open, visit this URL:\n%s\n\n", authURL)

	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Warning: Could not open browser automatically: %v\n", err)
		fmt.Println("Please open the URL manually in your browser.")
	}

	// Wait for authorization code with timeout
	var authCode string
	select {
	case code := <-codeChan:
		authCode = code
		fmt.Println("Authorization code received!")
	case err := <-errChan:
		fmt.Printf("Error from callback server: %v\n", err)
		fmt.Println("Falling back to manual authorization flow...")
		config.RedirectURL = originalRedirectURL
		return getTokenFromWebManual(config)
	case <-time.After(5 * time.Minute):
		fmt.Println("Timeout waiting for authorization. Falling back to manual flow...")
		config.RedirectURL = originalRedirectURL
		return getTokenFromWebManual(config)
	}

	// Exchange code for token
	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		fmt.Printf("Unable to retrieve token from web: %v\n", err)
		os.Exit(1)
	}

	return tok
}

// getTokenFromWebManual is the fallback manual authorization flow
func getTokenFromWebManual(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		fmt.Printf("Unable to read authorization code: %v", err)
		os.Exit(1)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		fmt.Printf("Unable to retrieve token from web: %v", err)
		os.Exit(1)
	}
	return tok
}

// savingTokenSource wraps an oauth2.TokenSource to save the token whenever it is refreshed.
type savingTokenSource struct {
	source oauth2.TokenSource
	path   string
	config *oauth2.Config
}

func (s *savingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := s.source.Token()
	if err != nil {
		return nil, err
	}

	// We only want to save if the token was actually refreshed.
	// oauth2.TokenSource might return the same token if it's still valid.
	// However, config.TokenSource often manages this.
	// To be safe and simple, we can load the current file and compare,
	// or just check if the expiry or access token is different from what we might have.
	// But since we want to be sure it's always up to date:

	current, _, _ := tokenFromFile(s.path)
	if current == nil || current.AccessToken != tok.AccessToken || !current.Expiry.Equal(tok.Expiry) {
		fmt.Printf("Token refreshed, saving to %s\n", s.path)
		saveToken(s.path, tok, s.config)
	}

	return tok, nil
}
