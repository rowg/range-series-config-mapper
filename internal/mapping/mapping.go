package mapping

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"git.axiom/axiom/hfradar-config-mapper/internal/config_interval"
)

const configDateTimePattern = `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z`

const configTimeLayout = "20060102T150405Z"
const configStartTimeIndex = 0
const configEndTimeIndex = 1

const operatorConfigTimeDelimiter = "-"

const rangeSeriesDateTimePattern = `\d{4}_\d{2}_\d{2}_\d{6}`
const rangeSeriesTimeLayout = "2006_01_02_150405"

func parseConfigDateTime(str string, pattern string) (time.Time, error) {
	// TODO: Refactor this function to be general purpose
	re, err := regexp.Compile(pattern)
	if err != nil {
		return time.Time{}, err
	}

	matches := re.FindStringSubmatch(str)
	if len(matches) == 1 {
		parsedTime, err := time.Parse(configTimeLayout, matches[0])
		if err != nil {
			return time.Time{}, err
		}
		return parsedTime, nil
	} else if len(matches) > 1 {
		return time.Time{}, fmt.Errorf("multiple matches found")
	}

	return time.Time{}, fmt.Errorf("no matches found")
}

func extractTimestampStr(str string, pattern string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(str)
	if len(matches) == 1 {
		return matches[0], nil
	} else if len(matches) > 1 {
		return "", fmt.Errorf("multiple matches found")
	}

	return "", fmt.Errorf("no matches found")
}

func BuildOperatorConfigIntervals(configs []string) []config_interval.ConfigInterval {
	res := []config_interval.ConfigInterval{}

	// Ensure configs are sorted
	slices.Sort(configs)

	for _, configPath := range configs {
		configFileName := filepath.Base(configPath)
		timeComponents := strings.Split(configFileName, operatorConfigTimeDelimiter)

		startTime, _ := parseConfigDateTime(timeComponents[configStartTimeIndex], configDateTimePattern)
		endTime, _ := parseConfigDateTime(timeComponents[configEndTimeIndex], configDateTimePattern)

		// Create new time interval
		timeInterval := config_interval.ConfigInterval{
			Start:  startTime,
			End:    endTime,
			Config: configPath,
		}
		res = append(res, timeInterval)
	}

	return res
}

func BuildAutoConfigIntervals(configs []string) []config_interval.ConfigInterval {
	res := []config_interval.ConfigInterval{}

	// Ensure configs are sorted
	slices.Sort(configs)

	for i, configPath := range configs {
		configTime, _ := parseConfigDateTime(configPath, configDateTimePattern)

		// Create new time interval
		timeInterval := config_interval.ConfigInterval{
			Start:  configTime,
			End:    time.Now().UTC().Truncate(time.Millisecond * 1000),
			Config: configPath,
		}
		res = append(res, timeInterval)

		// Set end time of previous interval to start time of current interval
		if i > 0 {
			res[i-1].End = configTime
		}
	}

	return res
}

func getMatchingConfig(timestamp time.Time, autoConfigTimeIntervals, operatorConfigTimeIntervals []config_interval.ConfigInterval) string {
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

	// TODO: What to do if no matching config?
	return ""
}

func CreateRangeSeriesToConfigMap(rangeSeriesFiles []string, autoConfigTimeIntervals, operatorConfigTimeIntervals []config_interval.ConfigInterval) map[string]string {
	fmt.Println("Computing RangeSeries:Config mapping...")

	result := make(map[string]string)

	// Iterate over each range series file
	for _, rangeSeriesPath := range rangeSeriesFiles {
		// 1. Extract base file name
		rangeSeriesName := filepath.Base(rangeSeriesPath)

		// 2. Parse timestamp from filename
		rangeSeriesTimeStr, _ := extractTimestampStr(rangeSeriesName, rangeSeriesDateTimePattern)
		rangeSeriesTime, _ := time.Parse(rangeSeriesTimeLayout, rangeSeriesTimeStr)

		// 3. Retrieve corresponding config file
		matchingConfig := getMatchingConfig(rangeSeriesTime, autoConfigTimeIntervals, operatorConfigTimeIntervals)

		// 4. Add key file path w/ value config file
		result[rangeSeriesPath] = matchingConfig
	}

	return result
}
