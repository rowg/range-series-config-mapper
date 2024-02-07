package main

import (
	"flag"
	"log"
	"path/filepath"

	"git.axiom/axiom/hfradar-config-mapper/internal/mapping"
	"git.axiom/axiom/hfradar-config-mapper/internal/read"
	"git.axiom/axiom/hfradar-config-mapper/internal/write"
)

const autoConfigDir = "Config_Auto"
const operatorConfigDir = "Config_Operator"
const rangeSeriesDir = "RangeSeries"
const configFileNamePattern = `\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}Z(\-\d{4}\d{2}\d{2}T\d{2}\d{2}\d{2}){0,1}$`
const rangeSeriesFilePathPattern = `\d{4}\/\d{2}\/\d{2}/.*.rs$`

func parseArgs() ([]string, bool, string, string, string) {
	// TODO: Name return values
	siteDir := flag.String("site-dir", "", "Absolute path to HFR site directory.")
	allRangeSeries := flag.Bool("all", false, "Boolean flag indicating whether to produce a mapping for all "+
		"RangeSeries files for the site. If set, `siteDir/RangeSeries` will be scanned for RangeSeries files.")
	outputFileType := flag.String("output-file-type", "JSON", "The format of the output file. Options are 'JSON' or 'CSV'.")
	outputFileName := flag.String("output-file-name", "rangeseries_to_config", "The name of the output file. Should not include the file ending.")

	flag.Parse()

	targetRangeSeriesFiles := flag.Args()

	log.Println("Target site directory:", *siteDir)
	log.Println("Output file type:", *outputFileType)
	log.Println("Output file name:", *outputFileName)
	log.Println("Targetting all RangeSeries files:", *allRangeSeries)

	return targetRangeSeriesFiles, *allRangeSeries, *siteDir, *outputFileType, *outputFileName
}

func validateArgs(targetRangeseriesFiles []string, allRangeSeries bool, siteDir, outputFileType, outputFileName string) {
	// siteDir must be specified
	if siteDir == "" {
		log.Fatalln("Error: --site-dir must be specified.")
	}

	// allRangeSeries and non-empty targetRangeseriesFiles are mutually-exclusive
	if allRangeSeries && len(targetRangeseriesFiles) > 0 {
		log.Fatalln("Error: Cannot specify individual RangeSeries files when the -all flag is active.")
	} else if !allRangeSeries && len(targetRangeseriesFiles) == 0 {
		log.Fatalln("Error: Must specify individual RangeSeries files when the -all flag is inactive.")
	}

	// outputFileType can only be `JSON` or `CSV`
	if !(outputFileType == "JSON" || outputFileType == "CSV") {
		log.Fatalf("Error: Invalid output-file-type of '%v'. Supported values are 'JSON' and 'CSV'.\n", outputFileType)
	}
}

func readConfigFiles(siteDir string, configType string) []string {
	log.Printf("Checking following path for configs: %v\n", filepath.Join(siteDir, configType))

	configPaths, err := read.FindFilesMatchingPattern(filepath.Join(siteDir, configType), configFileNamePattern, true)
	if err != nil {
		log.Fatalf("Error reading %s files: %v", configType, err)
	}

	return configPaths
}

func readRangeSeriesFiles(siteDir string) []string {
	log.Printf("Checking following path for RangeSeries files: %v\n", filepath.Join(siteDir, rangeSeriesDir))

	paths, err := read.FindFilesMatchingPattern(filepath.Join(siteDir, rangeSeriesDir), rangeSeriesFilePathPattern, false)
	if err != nil {
		log.Fatalf("Error reading RangeSeries files: %v", err)
	}

	return paths
}

func writeResult(mapping map[string]string, format string, fileName string) {
	log.Println("Writing mapping to disk...")

	if format == "JSON" {
		write.SaveMapAsJson(mapping, fileName)
	} else if format == "CSV" {
		write.SaveMapAsCsv(mapping, fileName)
	}
}

func main() {
	// 1. Parse CLI args
	targetRangeseriesFiles, allRangeSeries, siteDir, outputFileType, outputFileName := parseArgs()
	validateArgs(targetRangeseriesFiles, allRangeSeries, siteDir, outputFileType, outputFileName)

	// 2. Build mapping of time intervals to configs
	// 2.a. Retrieve configs
	autoConfigs := readConfigFiles(siteDir, autoConfigDir)
	operatorConfigs := readConfigFiles(siteDir, operatorConfigDir)

	// 2.b. Build mapping of time intervals to configs
	autoConfigIntervals := mapping.BuildAutoConfigIntervals(autoConfigs)
	operatorConfigIntervals := mapping.BuildOperatorConfigIntervals(operatorConfigs)

	// 3. Build mapping of RangeSeries files to Config directories
	var rangeSeriesFilePaths []string
	if allRangeSeries {
		rangeSeriesFilePaths = readRangeSeriesFiles(siteDir)
	} else {
		rangeSeriesFilePaths = targetRangeseriesFiles
	}

	rangeSeriesToConfig := mapping.CreateRangeSeriesToConfigMap(rangeSeriesFilePaths, autoConfigIntervals, operatorConfigIntervals)

	// 4. Write mapping to disk
	writeResult(rangeSeriesToConfig, outputFileType, outputFileName)

}
