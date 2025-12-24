package cleanup

import (
	"testing"
	"time"
)

func TestParseDatePattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "yyyy-MM-dd pattern",
			pattern:  "yyyy-MM-dd",
			expected: "2006-01-02",
		},
		{
			name:     "yyyyMMdd pattern",
			pattern:  "yyyyMMdd",
			expected: "20060102",
		},
		{
			name:     "yyyy/MM/dd pattern",
			pattern:  "yyyy/MM/dd",
			expected: "2006/01/02",
		},
		{
			name:     "dd-MM-yyyy pattern",
			pattern:  "dd-MM-yyyy",
			expected: "02-01-2006",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CleanupService{datePattern: tt.pattern}
			result := c.parseDatePattern(tt.pattern)
			if result != tt.expected {
				t.Errorf("parseDatePattern(%s) = %s, want %s", tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestMatchesDatePattern(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		folderName    string
		shouldMatch   bool
		expectedDate  string
	}{
		{
			name:         "matches yyyy-MM-dd",
			pattern:      "yyyy-MM-dd",
			folderName:   "2025-01-15",
			shouldMatch:  true,
			expectedDate: "2025-01-15",
		},
		{
			name:         "matches yyyyMMdd",
			pattern:      "yyyyMMdd",
			folderName:   "20250115",
			shouldMatch:  true,
			expectedDate: "2025-01-15",
		},
		{
			name:        "does not match - wrong format",
			pattern:     "yyyy-MM-dd",
			folderName:  "2025/01/15",
			shouldMatch: false,
		},
		{
			name:        "does not match - not a date",
			pattern:     "yyyy-MM-dd",
			folderName:  "Service A",
			shouldMatch: false,
		},
		{
			name:        "does not match - invalid date",
			pattern:     "yyyy-MM-dd",
			folderName:  "2025-13-45",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CleanupService{datePattern: tt.pattern}
			matches, parsedDate := c.matchesDatePattern(tt.folderName)

			if matches != tt.shouldMatch {
				t.Errorf("matchesDatePattern(%s) match = %v, want %v", tt.folderName, matches, tt.shouldMatch)
			}

			if tt.shouldMatch && matches {
				expectedTime, _ := time.Parse("2006-01-02", tt.expectedDate)
				if !parsedDate.Equal(expectedTime) {
					t.Errorf("matchesDatePattern(%s) date = %v, want %v", tt.folderName, parsedDate, expectedTime)
				}
			}
		})
	}
}
