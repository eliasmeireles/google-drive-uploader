package parser

import (
	"reflect"
	"testing"
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseFilename() = %v, want %v", got, tt.want)
			}
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
		if got := camelToSnakeCase(tt.input); got != tt.want {
			t.Errorf("camelToSnakeCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
