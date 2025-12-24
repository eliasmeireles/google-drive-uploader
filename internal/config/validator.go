package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	defaultConfigDir       = "/etc/google-drive-uploader"
	defaultTokenFile       = "token.json"
	defaultCredentialsFile = "client-secret.json"
)

var (
	DefaultTokenFilePath        = filepath.Join(defaultConfigDir, defaultTokenFile)
	DefaultCredentialsFilesPath = filepath.Join(defaultConfigDir, defaultCredentialsFile)
)

// Validate checks the configuration for errors and sets defaults
func (c *Config) Validate(args []string) error {
	// Handle token generation mode validation
	if c.TokenGen {
		if _, err := os.Stat(DefaultCredentialsFilesPath); err != nil {
			return err
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

	// Only require client-secret if we don't have a token and we are NOT in token-gen mode (already checked above)
	if !c.TokenGen {
		if _, err := os.Stat(c.TokenPath); err != nil {
			return err
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
