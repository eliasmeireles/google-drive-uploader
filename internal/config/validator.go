package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultConfigDir       = "/etc/google-driver-uploader"
	defaultCredentialsFile = "api-key.json"
	defaultTokenFile       = "token.json"
)

// Validate checks the configuration for errors and sets defaults
func (c *Config) Validate(args []string) error {
	// Handle token generation mode validation
	if c.TokenGen {
		if c.ClientSecret == "" {
			// Check default location
			defaultCreds := filepath.Join(defaultConfigDir, defaultCredentialsFile)
			if _, err := os.Stat(defaultCreds); err == nil {
				c.ClientSecret = defaultCreds
			} else {
				return fmt.Errorf("--client-secret is required for token generation")
			}
		}
		return nil
	}

	// Normal mode validation
	if c.RootFolderID == "" {
		return fmt.Errorf("--root-folder-id is required")
	}

	if len(args) == 0 && c.WorkDir == "" && !c.Cleanup {
		return fmt.Errorf("at least one file or --workdir is required (unless using --cleanup mode)")
	}

	// Handle default token path if not explicitly provided and local token file doesn't exist
	if c.TokenPath == ".out/token.json" {
		if _, err := os.Stat(".out/token.json"); os.IsNotExist(err) {
			defaultToken := filepath.Join(defaultConfigDir, defaultTokenFile)
			if _, err := os.Stat(defaultToken); err == nil {
				c.TokenPath = defaultToken
			}
		}
	}

	// Check if we have a valid token file
	hasValidToken := false
	if _, err := os.Stat(c.TokenPath); err == nil {
		hasValidToken = true
	}

	// Only require client-secret if we don't have a token and we are NOT in token-gen mode (already checked above)
	if !c.TokenGen && c.ClientSecret == "" {
		defaultCreds := filepath.Join(defaultConfigDir, defaultCredentialsFile)
		if _, err := os.Stat(defaultCreds); err == nil {
			c.ClientSecret = defaultCreds
		} else if !hasValidToken {
			// Check if token is self-sufficient (we can't easily check inside without parsing,
			// but authenticator will handle it. If it fails, it fails.)
			// We'll let authenticator try.
		}
	}

	// Validate cleanup-specific flags
	if c.Cleanup {
		if c.Keep < 1 {
			return fmt.Errorf("--keep must be at least 1")
		}
		if c.MatchPattern == "" {
			return fmt.Errorf("--match pattern is required for cleanup mode")
		}
	}

	return nil
}
