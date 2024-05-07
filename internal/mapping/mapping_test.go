package mapping

import (
	"reflect"
	"regexp"
	"testing"
	"time"

	"git.axiom/axiom/range-series-config-mapper/internal/config_interval"
	"git.axiom/axiom/range-series-config-mapper/internal/logger"
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

func TestBuildOperatorConfigIntervalsWithPresent(t *testing.T) {
	// Arrange
	configs := []string{
		"20230104T000000Z-present",
	}

	// Expected ConfigInterval fields
	expectedStart := time.Date(2023, 01, 04, 0, 0, 0, 0, time.UTC)
	expectedConfig := "20230104T000000Z-present"
	expectedEnd := time.Now().UTC()

	// Execute test
	got := BuildOperatorConfigIntervals(configs)

	// Assert
	if got[0].Start != expectedStart || got[0].Config != expectedConfig {
		t.Errorf("Mismatch in start or config: got %v", got[0])
	}

	// Check that the actual End is after the start of the test
	if got[0].End.Before(expectedEnd) {
		t.Errorf("End time %v is before the test start time %v", got[0].End, expectedEnd)
	}

	// Check that the actual end time is not unreasonably far in the future (more than a few seconds)
	if got[0].End.After(expectedEnd.Add(time.Second * 5)) {
		t.Errorf("End time %v is too far in the future compared to the test start time %v", got[0].End, expectedEnd)
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

func TestValidateOperatorConfigs(t *testing.T) {
	// Define test cases
	tests := []struct {
		name    string
		configs []config_interval.ConfigInterval
		wantErr bool
	}{
		{
			name: "Valid configs, single config",
			configs: []config_interval.ConfigInterval{
				{
					Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					Config: "20230101T000000Z-20230102T000000Z",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid configs, no gaps",
			configs: []config_interval.ConfigInterval{
				{
					Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					Config: "20230101T000000Z-20230102T000000Z",
				},
				{
					Start:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
					Config: "20230102T000000Z-20230103T000000Z",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid configs, gaps",
			configs: []config_interval.ConfigInterval{
				{
					Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					Config: "20230101T000000Z-20230102T000000Z",
				},
				{
					Start:  time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 7, 0, 0, 0, 0, time.UTC),
					Config: "20230103T000000Z-20230107T000000Z",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid configs, has `present`",
			configs: []config_interval.ConfigInterval{
				{
					Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					Config: "20230101T000000Z-20230102T000000Z",
				},
				{
					Start:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
					Config: "20230102T000000Z-20230103T000000Z",
				},
				{
					Start:  time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
					End:    time.Now().UTC(),
					Config: "20230103T000000Z-present",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid configs, overlapping",
			configs: []config_interval.ConfigInterval{
				{
					Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC),
					Config: "20230101T000000Z-20230104T000000Z",
				},
				{
					Start:  time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
					Config: "20230102T000000Z-20230103T000000Z",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid configs: overlapping with multiple `present` tokens",
			configs: []config_interval.ConfigInterval{
				{
					Start:  time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					End:    time.Now().UTC(),
					Config: "20230101T00000Z-present",
				},
				{
					Start:  time.Date(2023, 1, 5, 0, 0, 0, 0, time.UTC),
					End:    time.Now().UTC(),
					Config: "20230105T00000Z-present",
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid configs: start date in the future",
			configs: []config_interval.ConfigInterval{
				{
					Start:  time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC),
					End:    time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC),
					Config: "20250105T00000Z-20260105T000000Z",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &logger.TestLogger{}
			ValidateOperatorConfigs(tt.configs, logger)
			if logger.FatalCalled != tt.wantErr {
				t.Errorf("ValidateOperatorConfigs() error = %v, wantErr %v", logger.Logs, tt.wantErr)
			}
		})
	}
}
