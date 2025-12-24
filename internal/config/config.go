package config

// Config holds the configuration for the application
type Config struct {
	ClientSecret    string
	RootFolderID    string
	FileName        string
	FolderName      string
	TokenPath       string
	SmartOrganize   bool
	WorkDir         string
	DeleteOnSuccess bool
	DeleteOnDone    bool

	// Token generation mode
	TokenGen bool

	// Cleanup flags
	Cleanup      bool
	Keep         int
	MatchPattern string
}
