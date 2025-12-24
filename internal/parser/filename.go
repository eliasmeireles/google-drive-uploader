package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Metadata holds extracted information from the filename
type Metadata struct {
	Service string
	Date    string
}

// ParseFilename extracts metadata from a filename based on the pattern:
// [Service]_backup_[Date]_[Time]
// Example: oauth_backup_20251102_040000.sql.gz
// Valid Date formats: YYYYMMDD or YYYY-MM-DD
func ParseFilename(filename string) (*Metadata, error) {
	// Pattern explanation:
	// ^(?P<service>[a-zA-Z0-9]+)  : Start with alphanumeric service name
	// _backup_                    : Literal separator
	// (?P<date>\d{8}|\d{4}-\d{2}-\d{2}) : Date as 8 digits OR YYYY-MM-DD
	// _                           : Helper separator before time (time not strictly needed for folder but confirms pattern)
	// .*                          : Rest of the file
	re := regexp.MustCompile(`^(?P<service>[a-zA-Z0-9]+)_backup_(?P<date>\d{8}|\d{4}-\d{2}-\d{2})_.*`)

	matches := re.FindStringSubmatch(filename)
	if matches == nil {
		return nil, fmt.Errorf("filename '%s' does not match pattern '[service]_backup_[date]_...'", filename)
	}

	result := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = matches[i]
		}
	}

	return &Metadata{
		Service: camelToSnakeCase(result["service"]),
		Date:    normalizeDate(result["date"]),
	}, nil
}

// camelToSnakeCase converts a CamelCase string to SNAKE_CASE.
// Examples:
//   - "myAppService" -> "MY_APP_SERVICE"
//   - "OAuthBackup" -> "O_AUTH_BACKUP"
//   - "keycloak" -> "KEYCLOAK"
func camelToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		// Insert underscore before uppercase letters (except at the start)
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToUpper(result.String())
}

// normalizeDate takes a date string which might be YYYYMMDD or YYYY-MM-DD
// and returns it in YYYY-MM-DD format for consistency.
func normalizeDate(dateStr string) string {
	if strings.Contains(dateStr, "-") {
		return dateStr
	}
	// Assume YYYYMMDD if no dashes and validated by regex
	if len(dateStr) == 8 {
		return fmt.Sprintf("%s-%s-%s", dateStr[0:4], dateStr[4:6], dateStr[6:8])
	}
	return dateStr
}
