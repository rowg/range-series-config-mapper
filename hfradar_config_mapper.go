package main

import (
	"flag"
	"fmt"

	"git.axiom/axiom/hfradar-config-mapper/internal/mapping"
	"git.axiom/axiom/hfradar-config-mapper/internal/read"
	"git.axiom/axiom/hfradar-config-mapper/internal/write"
)

func parseArgs() (string, bool, string, string, []string) {
	// TODO: Arg validation
	// TODO: Single range series file
	// TODO: Multiple range series files
	// TODO: Add hints
	// TODO: Name return values
	targetSiteDir := flag.String("target-site-dir", "", "")
	allRangeSeries := flag.Bool("all", false, "")
	outputFileType := flag.String("output-file-type", "JSON", "")
	outputFileName := flag.String("output-file-name", "rangeseries_to_config", "")

	flag.Parse()

	targetRangeseriesFiles := flag.Args()

	fmt.Println("Target site directory:", *targetSiteDir)
	fmt.Println("All range series files:", *allRangeSeries)
	fmt.Println("Output file type:", *outputFileType)

	return *targetSiteDir, *allRangeSeries, *outputFileType, *outputFileName, targetRangeseriesFiles
}

func readConfigFiles(siteDir string, config_type string) ([]string, error) {
	config_paths, err := read.FindFilesMatchingPattern(siteDir+"/"+config_type, `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z$`, true)
	fmt.Printf("Checking following path for configs: %v\n", siteDir+"/"+config_type)

	if err != nil {
		return nil, err
	}

	return config_paths, err
}

func readRangeseriesFiles(siteDir string) ([]string, error) {
	paths, err := read.FindFilesMatchingPattern(siteDir+"/"+"RangeSeries", `\d{4}\/\d{2}\/\d{2}/.*.rs$`, false)

	if err != nil {
		return nil, err
	}

	return paths, err
}

func writeResult(mapping map[string]string, format string, fileName string) {
	// TODO: Handle neither format being passed in
	if format == "JSON" {
		write.SaveMapAsJson(mapping, fileName)
	} else if format == "CSV" {
		write.SaveMapAsCsv(mapping, fileName)
	}
}

func main() {
	// 1. Parse CLI args
	targetSiteDir, allRangeSeries, outputFileType, outputFileName, targetRangeseriesFiles := parseArgs()

	// 2. Build mapping of time intervals to configs
	// 2.a. Retrieve configs
	autoConfigs, _ := readConfigFiles(targetSiteDir, "Config_Auto")
	operatorConfigs, _ := readConfigFiles(targetSiteDir, "Config_Operator")

	// 2.b. Build mapping of time intervals to configs
	autoConfigIntervals := mapping.BuildAutoConfigIntervals(autoConfigs)
	operatorConfigIntervals := mapping.BuildOperatorConfigIntervals(operatorConfigs)

	// 3. Build mapping of RangeSeries files to Config directories
	var rangeSeriesFilePaths []string
	if allRangeSeries {
		rangeSeriesFilePaths, _ = readRangeseriesFiles(targetSiteDir)
	} else {
		rangeSeriesFilePaths = targetRangeseriesFiles
	}

	rangeSeriesToConfig := mapping.CreateRangeSeriesToConfigMap(rangeSeriesFilePaths, autoConfigIntervals, operatorConfigIntervals)

	// 4. Write mapping to disk
	writeResult(rangeSeriesToConfig, outputFileType, outputFileName)

}
