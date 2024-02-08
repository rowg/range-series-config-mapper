package mapping

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"git.axiom/axiom/hfradar-config-mapper/internal/config_interval"
)

func TestParseConfigDateTime(t *testing.T) {
	// Define test cases
	tests := []struct {
		name    string
		str     string
		pattern string
		want    time.Time
		wantErr bool
	}{
		{
			name:    "Valid date string",
			str:     "20230101T120000Z",
			pattern: `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z`,
			want:    time.Date(2023, 01, 01, 12, 00, 00, 0, time.UTC),
			wantErr: false,
		},
		{
			name:    "Invalid date string",
			str:     "20230101120000Z",
			pattern: `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z`,
			want:    time.Time{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseConfigDateTime(tt.str, tt.pattern)

			// Check if err matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("parseConfigDateTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if result matches expectation
			if !got.Equal(tt.want) && !tt.wantErr {
				t.Errorf("parseConfigDateTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTimestampStr(t *testing.T) {
	// Define test cases
	tests := []struct {
		name    string
		str     string
		pattern string
		want    string
		wantErr bool
	}{
		{
			name:    "Successful match",
			str:     "example_2023_01_01_120000.rs",
			pattern: `\d{4}_\d{2}_\d{2}_\d{6}`,
			want:    "2023_01_01_120000",
			wantErr: false,
		},
		{
			name:    "No match",
			str:     "example_20230101_120000.rs",
			pattern: `\d{4}_\d{2}_\d{2}_\d{6}`,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex, err := regexp.Compile(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to compile regex: %v", err)
			}

			got, err := extractTimestampStr(tt.str, regex)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTimestampStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractTimestampStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildOperatorConfigIntervals(t *testing.T) {
	// Arrange
	configs := []string{
		"20230101T000000Z-20230102T000000Z",
		"20230102T000000Z-20230103T000000Z",
	}

	expected := []config_interval.ConfigInterval{
		{
			Start:  time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC),
			End:    time.Date(2023, 01, 02, 0, 0, 0, 0, time.UTC),
			Config: "20230101T000000Z-20230102T000000Z",
		},
		{
			Start:  time.Date(2023, 01, 02, 0, 0, 0, 0, time.UTC),
			End:    time.Date(2023, 01, 03, 0, 0, 0, 0, time.UTC),
			Config: "20230102T000000Z-20230103T000000Z",
		},
	}

	// Execute test
	got := BuildOperatorConfigIntervals(configs)

	// Assert results
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("BuildOperatorConfigIntervals() = %v, want %v", got, expected)
	}
}

// Mock timeNow function for testing
var mockTimeNow = func() time.Time {
	return time.Date(2023, 01, 03, 0, 0, 0, 0, time.UTC)
}

func TestBuildAutoConfigIntervals(t *testing.T) {
	// Override the timeNow function in the test environment
	originalTimeNow := timeNow
	timeNow = mockTimeNow

	// Ensure we reset timeNow after the test
	defer func() { timeNow = originalTimeNow }()

	// Mock inputs
	configs := []string{
		"20230101T000000Z",
		"20230102T000000Z",
	}

	// Expected outputs
	expected := []config_interval.ConfigInterval{
		{
			Start:  time.Date(2023, 01, 01, 0, 0, 0, 0, time.UTC),
			End:    time.Date(2023, 01, 02, 0, 0, 0, 0, time.UTC),
			Config: "20230101T000000Z",
		},
		{
			Start:  time.Date(2023, 01, 02, 0, 0, 0, 0, time.UTC),
			End:    mockTimeNow(),
			Config: "20230102T000000Z",
		},
	}

	got := BuildAutoConfigIntervals(configs)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("BuildAutoConfigIntervals() = %v, want %v", got, expected)
	}
}
