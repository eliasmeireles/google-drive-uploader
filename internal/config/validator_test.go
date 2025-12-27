package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Validate_Upload(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create dummy client-secret.json and token.json
	apiKeyPath := filepath.Join(tempDir, "client-secret.json")
	os.WriteFile(apiKeyPath, []byte("{}"), 0600)
	tokenPath := filepath.Join(tempDir, "token.json")
	os.WriteFile(tokenPath, []byte("{}"), 0600)

	tests := []struct {
		name    string
		config  Config
		args    []string
		wantErr bool
	}{
		{
			name: "Valid upload config",
			config: Config{
				RootFolderID: "folder123",
				ClientSecret: apiKeyPath,
				TokenPath:    tokenPath,
			},
			args:    []string{"file.txt"},
			wantErr: false,
		},
		{
			name: "Valid upload config with workdir",
			config: Config{
				RootFolderID: "folder123",
				ClientSecret: apiKeyPath,
				TokenPath:    tokenPath,
				WorkDir:      "/tmp",
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "Missing root folder ID",
			config: Config{
				ClientSecret: apiKeyPath,
				TokenPath:    tokenPath,
			},
			args:    []string{"file.txt"},
			wantErr: true,
		},
		{
			name: "Missing files and workdir",
			config: Config{
				RootFolderID: "folder123",
				ClientSecret: apiKeyPath,
				TokenPath:    tokenPath,
			},
			args:    []string{},
			wantErr: true,
		},
		{
			name: "Valid cleanup config",
			config: Config{
				RootFolderID: "folder123",
				ClientSecret: apiKeyPath,
				TokenPath:    tokenPath,
				Cleanup:      true,
				Keep:         5,
				MatchPattern: "yyyy-MM-dd",
			},
			args:    []string{},
			wantErr: false,
		},
		{
			name: "Invalid cleanup config (keep < 1)",
			config: Config{
				RootFolderID: "folder123",
				ClientSecret: apiKeyPath,
				TokenPath:    tokenPath,
				Cleanup:      true,
				Keep:         0,
				MatchPattern: "yyyy-MM-dd",
			},
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validate_TokenGen(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	// Create dummy client-secret.json
	apiKeyPath := filepath.Join(tempDir, "client-secret.json")
	os.WriteFile(apiKeyPath, []byte("{}"), 0600)

	t.Run("Valid token-gen config", func(t *testing.T) {
		config := Config{
			TokenGen:     true,
			ClientSecret: apiKeyPath,
		}

		if err := config.Validate([]string{}); err != nil {
			t.Errorf("Config.Validate() error = %v, want nil", err)
		}
	})

	t.Run("Token-gen missing client secret", func(t *testing.T) {
		// Temporarily point to non-existent file
		nonExistentPath := filepath.Join(tempDir, "missing.json")

		config := Config{
			TokenGen:     true,
			ClientSecret: nonExistentPath,
		}
		if err := config.Validate([]string{}); err == nil {
			t.Error("Config.Validate() error = nil, want error")
		}
	})
}
