package config_interval

import (
	"testing"
	"time"
)

func TestConfigInterval_ContainsTime(t *testing.T) {
	// Setup a ConfigInterval for testing
	startTime := time.Date(2022, 10, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2022, 10, 31, 23, 59, 59, 0, time.UTC)
	testInterval := ConfigInterval{Start: startTime, End: endTime, Config: "20230101T120000Z"}

	// Define test cases
	tests := []struct {
		name      string
		timestamp time.Time
		want      bool
	}{
		{"Start Time", startTime, true},
		{"End Time", endTime, false}, // Note: End time is exclusive
		{"Within Interval", time.Date(2022, 10, 15, 12, 0, 0, 0, time.UTC), true},
		{"Before Interval", time.Date(2022, 9, 30, 23, 59, 59, 0, time.UTC), false},
		{"After Interval", time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC), false},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := testInterval.ContainsTime(tt.timestamp)
			if got != tt.want {
				t.Errorf("ConfigInterval.ContainsTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
