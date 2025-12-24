package main

import (
	"fmt"
	"os"

	"github.com/eliasferreira/google-drive-uploader/internal/app"
	"github.com/eliasferreira/google-drive-uploader/internal/config"

	"github.com/spf13/cobra"
)

func main() {
	var cfg config.Config

	var rootCmd = &cobra.Command{
		Use:   "uploader [files...]",
		Short: "Upload files to Google Drive",
		Long: `A high-performance CLI tool to upload files to Google Drive.
Supports large files, automatic folder organization, and resumable uploads.`,
		Args: cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := app.Run(cfg, args); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Flags
	rootCmd.Flags().StringVar(&cfg.ClientSecret, "client-secret", "", "Path to the OAuth 2.0 client secret file (optional if exists in /etc/google-drive-uploader/ or if token.json has embedded credentials)")
	rootCmd.Flags().StringVar(&cfg.RootFolderID, "root-folder-id", "", "ID of the root folder to save the file (required unless --token-gen is used)")
	rootCmd.Flags().StringVar(&cfg.FileName, "file-name", "", "Name of the file to save in Google Drive (optional, defaults to source filename). Note: Applied to ALL files if multiple.")
	rootCmd.Flags().StringVar(&cfg.FolderName, "folder-name", "", "Name of the sub-folder to save the file in (optional)")
	rootCmd.Flags().StringVar(&cfg.TokenPath, "token-path", ".out/token.json", "Path to the OAuth 2.0 token file (defaults to ./.out/token.json or /etc/google-drive-uploader/token.json)")
	rootCmd.Flags().BoolVar(&cfg.SmartOrganize, "smart-organize", false, "Enable smart organization based on filename")
	rootCmd.Flags().StringVar(&cfg.WorkDir, "workdir", "", "Path to the directory containing files to upload")
	rootCmd.Flags().BoolVar(&cfg.DeleteOnSuccess, "delete-on-success", false, "Delete the file after successful upload")
	rootCmd.Flags().BoolVar(&cfg.DeleteOnDone, "delete-on-done", false, "Delete the file after upload attempt (success or failure)")
	rootCmd.Flags().BoolVar(&cfg.TokenGen, "token-gen", false, "Generate token only (skips upload). Requires --client-secret")

	// Cleanup flags
	rootCmd.Flags().BoolVar(&cfg.Cleanup, "cleanup", false, "Enable cleanup mode to remove old date-based folders")
	rootCmd.Flags().IntVar(&cfg.Keep, "keep", 1, "Number of most recent date folders to keep (used with --cleanup)")
	rootCmd.Flags().StringVar(&cfg.MatchPattern, "match", "yyyy-MM-dd", "Date pattern to match folder names (e.g., yyyy-MM-dd, yyyyMMdd)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
