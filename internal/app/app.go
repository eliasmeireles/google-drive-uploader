package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/eliasferreira/google-drive-uploader/internal/auth"
	"github.com/eliasferreira/google-drive-uploader/internal/cleanup"
	"github.com/eliasferreira/google-drive-uploader/internal/config"
	"github.com/eliasferreira/google-drive-uploader/internal/driveclient"
	"github.com/eliasferreira/google-drive-uploader/internal/parser"
)

// Run executes the main application using the provided configuration
func Run(cfg config.Config, args []string) error {
	ctx := context.Background()

	// 1. Validate Config
	if err := cfg.Validate(args); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// 2. Authentication
	authenticator := auth.NewAuthenticator(cfg)
	client, err := authenticator.GetClient(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// If in token generation mode, we are done
	if cfg.TokenGen {
		fmt.Printf("Token successfully generated and saved to: %s\n", cfg.TokenPath)
		return nil
	}

	// 3. Drive Service (Once for all files)
	svc, err := driveclient.NewDriveService(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to create drive service: %w", err)
	}

	// 4. Check if cleanup mode is enabled
	if cfg.Cleanup {
		return runCleanup(ctx, svc, cfg)
	}

	// 5. Normal file upload
	return runUploads(ctx, svc, cfg, args)
}

func runCleanup(ctx context.Context, svc *driveclient.DriveService, cfg config.Config) error {
	// Run cleanup
	cleanupSvc := cleanup.NewCleanupService(svc, cfg.MatchPattern, cfg.Keep)
	deletedPaths, err := cleanupSvc.Run(ctx, cfg.RootFolderID)
	if err != nil {
		return fmt.Errorf("cleanup failed: %w", err)
	}

	// Log all deleted paths
	if len(deletedPaths) > 0 {
		fmt.Println("\n=== Deleted Folders ===")
		for _, path := range deletedPaths {
			fmt.Printf("  - %s\n", path)
		}
	} else {
		fmt.Println("No folders were deleted.")
	}

	return nil
}

func runUploads(ctx context.Context, svc *driveclient.DriveService, cfg config.Config, args []string) error {
	filesToProcess := args
	if cfg.WorkDir != "" {
		entries, err := os.ReadDir(cfg.WorkDir)
		if err != nil {
			return fmt.Errorf("failed to read workdir: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				filesToProcess = append(filesToProcess, filepath.Join(cfg.WorkDir, entry.Name()))
			}
		}
	}

	// Validate --file-name usage with multiple files
	if len(filesToProcess) > 1 && cfg.FileName != "" {
		fmt.Println("Warning: --file-name is ignored because multiple files were provided. Using original filenames.")
		cfg.FileName = ""
	}

	// Process each file
	for _, filePath := range filesToProcess {
		processFile(ctx, svc, cfg, filePath)
	}

	return nil
}

func processFile(ctx context.Context, svc *driveclient.DriveService, cfg config.Config, filePath string) {
	fmt.Printf("\n--- Processing: %s ---\n", filePath)

	// Basic validation
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Printf("Error: File '%s' does not exist. Skipping.", filePath)
		return
	}
	if info.IsDir() {
		log.Printf("Error: '%s' is a directory. Skipping.", filePath)
		return
	}

	// Determine Filename
	targetFileName := cfg.FileName
	if targetFileName == "" {
		targetFileName = filepath.Base(filePath)
	}

	// Handle Directory Logic
	parentID := cfg.RootFolderID

	// 1. Explicit Folder Name
	if cfg.FolderName != "" {
		id, err := svc.FindOrCreateFolder(ctx, cfg.FolderName, parentID)
		if err != nil {
			log.Printf("Failed to find or create folder '%s': %v. Skipping file.", cfg.FolderName, err)
			return
		}
		parentID = id
	}

	// 2. Smart Organization Logic
	if cfg.SmartOrganize {
		meta, err := parser.ParseFilename(targetFileName)
		if err != nil {
			fmt.Printf("Warning: Could not parse filename for smart organization: %v. Proceeding in current folder.\n", err)
		} else {
			fmt.Printf("Smart Organize: Service='%s', Date='%s'\n", meta.Service, meta.Date)

			// Service Folder
			sID, err := svc.FindOrCreateFolder(ctx, meta.Service, parentID)
			if err != nil {
				log.Printf("Failed to create Service folder: %v. Skipping file.", err)
				return
			}
			parentID = sID

			// Date Folder
			dID, err := svc.FindOrCreateFolder(ctx, meta.Date, parentID)
			if err != nil {
				log.Printf("Failed to create Date folder: %v. Skipping file.", err)
				return
			}
			parentID = dID
		}
	}

	// Upload
	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("Failed to open file: %v. Skipping.", err)
		return
	}
	defer f.Close()

	fmt.Printf("Uploading as '%s' to folder ID '%s'...\n", targetFileName, parentID)
	file, err := svc.UploadFile(ctx, f, targetFileName, parentID)
	if err != nil {
		log.Printf("Upload failed: %v", err)
		if cfg.DeleteOnDone {
			fmt.Printf("Removing file after failure: %s\n", filePath)
			os.Remove(filePath)
		}
		return
	}

	fmt.Printf("Success! ID: %s, Size: %d bytes\n", file.Id, file.Size)

	if cfg.DeleteOnSuccess || cfg.DeleteOnDone {
		fmt.Printf("Removing file after success: %s\n", filePath)
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Failed to remove file: %v", err)
		}
	}
}
