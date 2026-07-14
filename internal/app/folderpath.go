package app

import (
	"context"
	"strings"
)

// FolderCreator resolves a folder by name under a parent, creating it when absent.
type FolderCreator interface {
	FindOrCreateFolder(ctx context.Context, name string, parentID string) (string, error)
}

// ResolveFolderPath walks a "/"-separated folder path (e.g. "ORACLE-VPS-01/MONGODB"),
// finding or creating each segment under the previous one, and returns the ID of the
// deepest folder. Empty or blank segments are ignored, so "A//B" and "/A/B/" behave
// like "A/B". Returns parentID unchanged when the path has no usable segments.
func ResolveFolderPath(ctx context.Context, fc FolderCreator, folderPath string, parentID string) (string, error) {
	for segment := range strings.SplitSeq(folderPath, "/") {
		segment = strings.TrimSpace(segment)
		if segment == "" {
			continue
		}
		id, err := fc.FindOrCreateFolder(ctx, segment, parentID)
		if err != nil {
			return "", err
		}
		parentID = id
	}
	return parentID, nil
}
