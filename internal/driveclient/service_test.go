package driveclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func newTestDriveService(t *testing.T, handler http.HandlerFunc) *DriveService {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	srv, err := drive.NewService(context.Background(),
		option.WithHTTPClient(server.Client()),
		option.WithEndpoint(server.URL+"/"),
	)
	require.NoError(t, err)

	return &DriveService{srv: srv}
}

func TestUploadFile(t *testing.T) {
	t.Run("given a successful upload then returns the size reported by drive", func(t *testing.T) {
		var requestedFields string
		var receivedBody string
		svc := newTestDriveService(t, func(w http.ResponseWriter, r *http.Request) {
			requestedFields = r.URL.Query().Get("fields")
			body, _ := io.ReadAll(r.Body)
			receivedBody = string(body)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"file-123","name":"backup.sql.gz","size":"7130317"}`))
		})

		file, err := svc.UploadFile(context.Background(), strings.NewReader("backup-content"), "backup.sql.gz", "parent-1")

		require.NoError(t, err)
		assert.Equal(t, uploadResponseFields, requestedFields)
		assert.Contains(t, receivedBody, "backup-content")
		assert.Equal(t, "file-123", file.Id)
		assert.Equal(t, "backup.sql.gz", file.Name)
		assert.Equal(t, int64(7130317), file.Size)
	})

	t.Run("given drive returns an error then returns a wrapped error", func(t *testing.T) {
		svc := newTestDriveService(t, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":{"code":403,"message":"insufficient permissions"}}`))
		})

		file, err := svc.UploadFile(context.Background(), strings.NewReader("backup-content"), "backup.sql.gz", "parent-1")

		require.Error(t, err)
		assert.Nil(t, file)
		assert.Contains(t, err.Error(), "could not upload file")
	})
}
