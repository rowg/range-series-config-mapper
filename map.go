package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

func parseDateTimeWithRegex(str string, pattern string) (time.Time, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return time.Time{}, err
	}

	matches := re.FindStringSubmatch(str)
	if len(matches) > 0 {
		parsedTime, err := time.Parse("20060102T150405Z", matches[0])
		if err != nil {
			return time.Time{}, err
		}
		return parsedTime, nil
	}
	return time.Time{}, fmt.Errorf("no matches found")
}

func extractTimestampStr(str string, pattern string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(str)
	if len(matches) > 0 {
		return matches[0], nil
	}

	return "", fmt.Errorf("no matches found")
}

func buildOperatorConfigIntervals(configs []string) []ConfigInterval {
	res := []ConfigInterval{}

	// Ensure configs are sorted
	slices.Sort(configs)

	for _, configPath := range configs {
		configFileName := filepath.Base(configPath)
		timeComponents := strings.Split(configFileName, "-")

		startTime, _ := parseDateTimeWithRegex(timeComponents[0], `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z`)
		endTime, _ := parseDateTimeWithRegex(timeComponents[1], `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z`)

		// Create new time interval
		timeInterval := ConfigInterval{startTime, endTime, configPath}
		res = append(res, timeInterval)
	}

	return res
}

func buildAutoConfigIntervals(configs []string) []ConfigInterval {
	res := []ConfigInterval{}

	// Ensure configs are sorted
	slices.Sort(configs)

	for i, configPath := range configs {
		configTime, _ := parseDateTimeWithRegex(configPath, `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z`)

		// Create new time interval
		timeInterval := ConfigInterval{configTime, time.Now().UTC().Truncate(time.Millisecond * 1000), configPath}
		res = append(res, timeInterval)

		// Set end time of previous interval to start time of current interval
		if i > 0 {
			res[i-1].End = configTime
		}
	}

	return res
}

func getMatchingConfig(timestamp time.Time, autoConfigTimeIntervals, operatorConfigTimeIntervals []ConfigInterval) string {
	for _, timeInterval := range operatorConfigTimeIntervals {
		if timeInterval.ContainsTime(timestamp) {
			return timeInterval.Config
		}
	}

	for _, timeInterval := range autoConfigTimeIntervals {
		if timeInterval.ContainsTime(timestamp) {
			return timeInterval.Config
		}
	}

	// What to do if no matching config?
	return ""
}

func createRangeSeriesToConfigMap(rangeSeriesFiles []string, autoConfigTimeIntervals, operatorConfigTimeIntervals []ConfigInterval) map[string]string {
	result := make(map[string]string)

	// Iterate over each range series file
	for _, rangeSeriesPath := range rangeSeriesFiles {
		// 1. Extract base file name
		rangeSeriesName := filepath.Base(rangeSeriesPath)

		// 2. Parse timestamp from filename
		rangeSeriesTimeStr, _ := extractTimestampStr(rangeSeriesName, `\d{4}_\d{2}_\d{2}_\d{6}`)
		rangeSeriesTime, _ := time.Parse("2006_01_02_150405", rangeSeriesTimeStr)

		// 3. Retrieve corresponding config file
		matchingConfig := getMatchingConfig(rangeSeriesTime, autoConfigTimeIntervals, operatorConfigTimeIntervals)

		// 4. Add key file path w/ value config file
		result[rangeSeriesPath] = matchingConfig
	}

	return result
}
