package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTokenFromFile(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Legacy Token Format", func(t *testing.T) {
		tokenPath := filepath.Join(tempDir, "legacy_token.json")
		legacyToken := oauth2.Token{
			AccessToken:  "legacy-access-token",
			TokenType:    "Bearer",
			RefreshToken: "legacy-refresh-token",
			Expiry:       time.Now().Add(1 * time.Hour),
		}

		// Write legacy token (just the oauth2.Token struct)
		f, _ := os.Create(tokenPath)
		json.NewEncoder(f).Encode(legacyToken)
		f.Close()

		tok, tokenFile, err := tokenFromFile(tokenPath)
		if err != nil {
			t.Fatalf("tokenFromFile() error = %v", err)
		}

		if tok.AccessToken != legacyToken.AccessToken {
			t.Errorf("got AccessToken = %v, want %v", tok.AccessToken, legacyToken.AccessToken)
		}
		// In legacy format, TokenFile struct fields might be empty or zero-value if not matched,
		// but since TokenFile fields match oauth2.Token json tags mostly (except different variable names for json keys)
		// Wait, TokenFile has explict json tags: access_token, etc.
		// oauth2.Token also uses access_token, etc. So it should decode correctly into TokenFile too!
		if tokenFile.AccessToken != legacyToken.AccessToken {
			t.Errorf("TokenFile access token mismatch. got %v, want %v", tokenFile.AccessToken, legacyToken.AccessToken)
		}
		if tokenFile.ClientID != "" {
			t.Errorf("Expected empty ClientID for legacy token, got %v", tokenFile.ClientID)
		}
	})

	t.Run("Enhanced Token Format", func(t *testing.T) {
		tokenPath := filepath.Join(tempDir, "enhanced_token.json")
		enhancedToken := TokenFile{
			AccessToken:  "enhanced-access-token",
			TokenType:    "Bearer",
			RefreshToken: "enhanced-refresh-token",
			Expiry:       time.Now().Add(1 * time.Hour),
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		f, _ := os.Create(tokenPath)
		json.NewEncoder(f).Encode(enhancedToken)
		f.Close()

		tok, tokenFile, err := tokenFromFile(tokenPath)
		if err != nil {
			t.Fatalf("tokenFromFile() error = %v", err)
		}

		if tok.AccessToken != enhancedToken.AccessToken {
			t.Errorf("got AccessToken = %v, want %v", tok.AccessToken, enhancedToken.AccessToken)
		}
		if tokenFile.ClientID != enhancedToken.ClientID {
			t.Errorf("got ClientID = %v, want %v", tokenFile.ClientID, enhancedToken.ClientID)
		}
	})

	t.Run("File Not Found", func(t *testing.T) {
		_, _, err := tokenFromFile("non-existent-file.json")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

func TestSaveToken(t *testing.T) {
	tempDir := t.TempDir()
	tokenPath := filepath.Join(tempDir, "saved_token.json")

	token := &oauth2.Token{
		AccessToken:  "access-token",
		TokenType:    "Bearer",
		RefreshToken: "refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	config := &oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	}

	tokenData := NewTokenFile(config.ClientID, config.ClientSecret)
	tokenData.Refresh(token)

	// Save with config (enhanced)
	saveToken(tokenPath, tokenData)

	// Read back
	_, tokenFile, err := tokenFromFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read back saved token: %v", err)
	}

	if tokenFile.AccessToken != token.AccessToken {
		t.Errorf("Saved AccessToken mismatch")
	}
	if tokenFile.ClientID != config.ClientID {
		t.Errorf("Saved ClientID mismatch. Got %s, Want %s", tokenFile.ClientID, config.ClientID)
	}
}
