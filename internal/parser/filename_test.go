package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     *Metadata
		wantErr  bool
	}{
		{
			name:     "Valid Format YYYYMMDD",
			filename: "oauth_backup_20251102_040000.sql.gz",
			want: &Metadata{
				Service: "OAUTH",
				Date:    "2025-11-02",
			},
			wantErr: false,
		},
		{
			name:     "Valid Format YYYY-MM-DD",
			filename: "keycloak_backup_2025-11-02_040000.sql",
			want: &Metadata{
				Service: "KEYCLOAK",
				Date:    "2025-11-02",
			},
			wantErr: false,
		},
		{
			name:     "CamelCase Service Name",
			filename: "MyAppService_backup_20251102_040000.zip",
			want: &Metadata{
				Service: "MY_APP_SERVICE",
				Date:    "2025-11-02",
			},
			wantErr: false,
		},
		{
			name:     "CamelCase Service Name 2",
			filename: "OAuthBackup_backup_20251102_040000.zip",
			want: &Metadata{
				Service: "O_AUTH_BACKUP", // Based on current implementation logic of insert underscore before Upper
				Date:    "2025-11-02",
			},
			wantErr: false,
		},
		{
			name:     "Underscore Service Name YYYYMMDD",
			filename: "app_service_data_backup_20260324_164627.tar.gz",
			want: &Metadata{
				Service: "APP_SERVICE_DATA",
				Date:    "2026-03-24",
			},
			wantErr: false,
		},
		{
			name:     "Underscore Service Name YYYY-MM-DD",
			filename: "order_items_backup_2025-12-24_084205.tar.gz",
			want: &Metadata{
				Service: "ORDER_ITEMS",
				Date:    "2025-12-24",
			},
			wantErr: false,
		},
		{
			name:     "Invalid Pattern",
			filename: "random_file.txt",
			want:     nil,
			wantErr:  true,
		},
		{
			name:     "Missing Date",
			filename: "service_backup_no_date.zip",
			want:     nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilename(tt.filename)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCamelToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"myApp", "MY_APP"},
		{"simple", "SIMPLE"},
		{"OAuth", "O_AUTH"},
		{"APIClient", "A_P_I_CLIENT"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, camelToSnakeCase(tt.input))
	}
}
