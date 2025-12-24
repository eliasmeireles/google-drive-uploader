package cleanup

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
)

// DriveService defines the interface for Drive operations needed by cleanup
type DriveService interface {
	ListFolders(ctx context.Context, parentID string) ([]*drive.File, error)
	TrashFile(ctx context.Context, fileID string) error
}

// CleanupService handles cleanup operations
type CleanupService struct {
	driveService DriveService
	datePattern  string
	keepCount    int
	deletedPaths []string
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(driveService DriveService, datePattern string, keepCount int) *CleanupService {
	return &CleanupService{
		driveService: driveService,
		datePattern:  datePattern,
		keepCount:    keepCount,
		deletedPaths: make([]string, 0),
	}
}

// Run executes the cleanup process starting from rootFolderID
func (c *CleanupService) Run(ctx context.Context, rootFolderID string) ([]string, error) {
	log.Printf("Starting cleanup with pattern '%s', keeping %d most recent folders", c.datePattern, c.keepCount)

	err := c.traverseFolders(ctx, rootFolderID, "")
	if err != nil {
		return nil, err
	}

	log.Printf("Cleanup completed. Total folders moved to trash: %d", len(c.deletedPaths))
	return c.deletedPaths, nil
}

// FolderWithDate holds a folder and its parsed date
type FolderWithDate struct {
	File *drive.File
	Date time.Time
}

// traverseFolders recursively traverses folders and applies retention policy
func (c *CleanupService) traverseFolders(ctx context.Context, folderID string, currentPath string) error {
	folders, err := c.driveService.ListFolders(ctx, folderID)
	if err != nil {
		return fmt.Errorf("failed to list folders in '%s': %v", currentPath, err)
	}

	if len(folders) == 0 {
		return nil
	}

	// Separate folders into date-matching and non-date-matching
	var dateFolders []FolderWithDate
	var nonDateFolders []*drive.File

	for _, folder := range folders {
		matches, parsedDate := c.matchesDatePattern(folder.Name)
		if matches {
			dateFolders = append(dateFolders, FolderWithDate{
				File: folder,
				Date: parsedDate,
			})
		} else {
			nonDateFolders = append(nonDateFolders, folder)
		}
	}

	// If we have multiple date-matching folders, apply retention policy
	if len(dateFolders) > c.keepCount {
		err := c.applyRetentionPolicy(ctx, dateFolders, currentPath)
		if err != nil {
			return err
		}
	}

	// Recursively traverse non-date folders
	for _, folder := range nonDateFolders {
		newPath := currentPath
		if newPath == "" {
			newPath = folder.Name
		} else {
			newPath = currentPath + "/" + folder.Name
		}
		err := c.traverseFolders(ctx, folder.Id, newPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// applyRetentionPolicy sorts folders by date and keeps only the most recent ones
func (c *CleanupService) applyRetentionPolicy(ctx context.Context, folders []FolderWithDate, parentPath string) error {
	// Sort by date descending (most recent first)
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].Date.After(folders[j].Date)
	})

	// Keep the first keepCount folders, trash the rest
	for i := c.keepCount; i < len(folders); i++ {
		folder := folders[i].File
		fullPath := parentPath
		if fullPath == "" {
			fullPath = folder.Name
		} else {
			fullPath = parentPath + "/" + folder.Name
		}

		log.Printf("Moving to trash: %s (date: %s)", fullPath, folders[i].Date.Format("2006-01-02"))

		err := c.driveService.TrashFile(ctx, folder.Id)
		if err != nil {
			log.Printf("Warning: Failed to trash folder '%s': %v", fullPath, err)
			continue
		}

		c.deletedPaths = append(c.deletedPaths, fullPath)
	}

	return nil
}

// matchesDatePattern checks if a folder name matches the date pattern
func (c *CleanupService) matchesDatePattern(name string) (bool, time.Time) {
	goPattern := c.parseDatePattern(c.datePattern)
	if goPattern == "" {
		return false, time.Time{}
	}

	parsedDate, err := time.Parse(goPattern, name)
	if err != nil {
		return false, time.Time{}
	}

	return true, parsedDate
}

// parseDatePattern converts user-friendly date patterns to Go time format
func (c *CleanupService) parseDatePattern(pattern string) string {
	// Replace longer patterns first to avoid incorrect substring replacements
	// For example, "yyyy" must be replaced before "yy"
	result := pattern
	result = strings.ReplaceAll(result, "yyyy", "2006")
	result = strings.ReplaceAll(result, "yy", "06")
	result = strings.ReplaceAll(result, "MM", "01")
	result = strings.ReplaceAll(result, "dd", "02")

	return result
}
