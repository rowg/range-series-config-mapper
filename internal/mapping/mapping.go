package mapping

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"git.axiom/axiom/range-series-config-mapper/internal/config_interval"
	"git.axiom/axiom/range-series-config-mapper/internal/logger"
)

const (
	configDateTimePattern = `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z`
	configTimeLayout      = "20060102T150405Z"
	configStartTimeIndex  = 0
	configEndTimeIndex    = 1
)

const (
	operatorConfigTimeDelimiter = "-"
	presentToken                = "present"
)

const (
	rangeSeriesDateTimePattern = `\d{4}_\d{2}_\d{2}_\d{6}`
	rangeSeriesTimeLayout      = "2006_01_02_150405"
)

var timeNow = func() time.Time {
	return time.Now()
}

func parseConfigDateTime(str string, pattern string) (time.Time, error) {
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
	}

	return time.Time{}, fmt.Errorf("no matches found")
}

func extractTimestampStr(str string, regex *regexp.Regexp) (string, error) {
	matches := regex.FindStringSubmatch(str)
	if len(matches) == 1 {
		return matches[0], nil
	}

	return "", fmt.Errorf("no matches found")
}

func BuildOperatorConfigIntervals(configs []string) []config_interval.ConfigInterval {
	res := []config_interval.ConfigInterval{}

	configs = slices.Clone(configs)

	// Ensure configs are sorted
	slices.SortFunc(configs, func(a, b string) int {
		// Standard string sort, except when `present` is in the string.
		// `present` should be considered greater than any valid timestamp
		if strings.Contains(a, presentToken) && !strings.Contains(b, presentToken) {
			return 1
		} else if !strings.Contains(a, presentToken) && strings.Contains(b, presentToken) {
			return -1
		}

		// If both or neither timestamp contains present, sort as regular strings
		return strings.Compare(a, b)
	})

	for _, configPath := range configs {
		configFileName := filepath.Base(configPath)
		timeComponents := strings.Split(configFileName, operatorConfigTimeDelimiter)

		startTime, err := parseConfigDateTime(timeComponents[configStartTimeIndex], configDateTimePattern)
		if err != nil {
			log.Fatalf("Error parsing Operator Config start time: %v", err)
		}

		var endTime time.Time
		// If the end time component is presentToken, assign endTime to the current UTC time
		if timeComponents[configEndTimeIndex] == presentToken {
			endTime = time.Now().UTC()
		} else {
			endTime, err = parseConfigDateTime(timeComponents[configEndTimeIndex], configDateTimePattern)
			if err != nil {
				log.Fatalf("Error parsing Operator Config end time: %v", err)
			}
		}

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

func ValidateOperatorConfigs(configs []config_interval.ConfigInterval, logger logger.Logger) {
	for i, config := range configs {
		// Ensure configs do not overlap
		if i > 0 && config.Start.Before(configs[i-1].End) {
			logger.Fatalf("Error: Operator config %v overlaps with %v", configs[i-1].Config, config.Config)
		}

		// Ensure that configs are not in the future
		if time.Now().Before(config.Start) {
			logger.Fatalf("Error: Operator config %v is in the future", config.Config)
		}
	}
}

func BuildAutoConfigIntervals(configs []string) []config_interval.ConfigInterval {
	res := []config_interval.ConfigInterval{}

	// Ensure configs are sorted
	slices.Sort(configs)

	for i, configPath := range configs {
		configTime, err := parseConfigDateTime(configPath, configDateTimePattern)
		if err != nil {
			log.Fatalf("Error parsing Autodetected Config start time: %v", err)
		}

		// Create new time interval
		timeInterval := config_interval.ConfigInterval{
			Start:  configTime,
			End:    timeNow().UTC().Truncate(time.Millisecond * 1000),
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

	// Return an empty string is there is no matching config
	return ""
}

func CreateRangeSeriesToConfigMap(rangeSeriesFiles []string, autoConfigTimeIntervals, operatorConfigTimeIntervals []config_interval.ConfigInterval) map[string]string {
	log.Println("Computing RangeSeries:Config mapping...")

	result := make(map[string]string)

	rangeSeriesDateTimeRegex, err := regexp.Compile(rangeSeriesDateTimePattern)
	if err != nil {
		log.Fatalf("Error compiling rangeSeriesDateTimeRegex: %v", err)
	}

	// Iterate over each range series file
	for _, rangeSeriesPath := range rangeSeriesFiles {
		// 1. Extract base file name
		rangeSeriesName := filepath.Base(rangeSeriesPath)

		// 2. Parse timestamp from filename
		rangeSeriesTimeStr, err := extractTimestampStr(rangeSeriesName, rangeSeriesDateTimeRegex)
		if err != nil {
			log.Printf("Skipping RangeSeries file '%s': problem extracting timestamp from filename: %v\n", rangeSeriesName, err)
			continue
		}

		rangeSeriesTime, err := time.Parse(rangeSeriesTimeLayout, rangeSeriesTimeStr)
		if err != nil {
			log.Printf("Skipping RangeSeries file '%s': problem parsing filename timestamp: %v\n", rangeSeriesName, err)
			continue
		}

		// 3. Retrieve corresponding config file
		matchingConfig := getMatchingConfig(rangeSeriesTime, autoConfigTimeIntervals, operatorConfigTimeIntervals)

		// 4. Add key file path w/ value config file
		result[rangeSeriesPath] = matchingConfig
	}

	return result
}
