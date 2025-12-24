package driveclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// Service defines the interface for interacting with Google Drive
type Service interface {
	UploadFile(ctx context.Context, file io.Reader, filename string, parentID string) (*drive.File, error)
	FindOrCreateFolder(ctx context.Context, name string, parentID string) (string, error)
}

// DriveService implements the Service interface for Google Drive
type DriveService struct {
	srv *drive.Service
}

// NewDriveService creates a new DriveService
func NewDriveService(ctx context.Context, client *http.Client) (*DriveService, error) {
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Drive client: %v", err)
	}

	return &DriveService{srv: srv}, nil
}

// UploadFile uploads a file to Google Drive.
// It uses Resumable Uploads which is good for large files.
func (s *DriveService) UploadFile(ctx context.Context, file io.Reader, filename string, parentID string) (*drive.File, error) {
	f := &drive.File{
		Name:    filename,
		Parents: []string{parentID},
	}

	// We can add Progress reporting if needed by wrapping the reader,
	// but for now we stick to the basic resumable upload.
	res, err := s.srv.Files.Create(f).Media(file).Do()
	if err != nil {
		return nil, fmt.Errorf("could not upload file: %v", err)
	}

	return res, nil
}

// FindOrCreateFolder checks if a folder exists with the given name in the parentID.
// If it exists, returns its ID. If not, creates it and returns the new ID.
func (s *DriveService) FindOrCreateFolder(ctx context.Context, name string, parentID string) (string, error) {
	// 1. Search for the folder
	q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents and trashed = false", name, parentID)

	// Using simple string concatenation for query is prone to injection if inputs aren't sanitized,
	// but here we trust the inputs or sanitize them.
	// A better approach is to rely on the API but it uses string queries.
	// For this CLI tool, we'll assume basic usage.

	// Escaping single quotes in name to prevent errors:
	escapedName := strings.ReplaceAll(name, "'", "\\'")
	q = fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and name = '%s' and '%s' in parents and trashed = false", escapedName, parentID)

	r, err := s.srv.Files.List().PageSize(1).Q(q).Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve files: %v", err)
	}

	if len(r.Files) > 0 {
		return r.Files[0].Id, nil
	}

	// 2. Create if not found
	f := &drive.File{
		Name:     name,
		MimeType: "application/vnd.google-apps.folder",
		Parents:  []string{parentID},
	}

	res, err := s.srv.Files.Create(f).Do()
	if err != nil {
		return "", fmt.Errorf("could not create folder: %v", err)
	}

	return res.Id, nil
}

// ListFolders lists all folders within a parent folder
func (s *DriveService) ListFolders(ctx context.Context, parentID string) ([]*drive.File, error) {
	q := fmt.Sprintf("mimeType = 'application/vnd.google-apps.folder' and '%s' in parents and trashed = false", parentID)

	var allFolders []*drive.File
	pageToken := ""

	for {
		call := s.srv.Files.List().
			PageSize(100).
			Q(q).
			Fields("nextPageToken, files(id, name)")

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		r, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve folders: %v", err)
		}

		allFolders = append(allFolders, r.Files...)

		pageToken = r.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return allFolders, nil
}

// TrashFile moves a file or folder to trash
func (s *DriveService) TrashFile(ctx context.Context, fileID string) error {
	_, err := s.srv.Files.Update(fileID, &drive.File{
		Trashed: true,
	}).Do()

	if err != nil {
		return fmt.Errorf("could not trash file: %v", err)
	}

	return nil
}
