package main

import (
	"flag"
	"fmt"
	"log"

	"git.axiom/axiom/hfradar-config-mapper/internal/mapping"
	"git.axiom/axiom/hfradar-config-mapper/internal/read"
	"git.axiom/axiom/hfradar-config-mapper/internal/write"
)

func parseArgs() ([]string, bool, string, string, string) {
	// TODO: Name return values
	siteDir := flag.String("site-dir", "", "Absolute path to HFR site directory.")
	allRangeSeries := flag.Bool("all", false, "Boolean flag indicating whether to produce a mapping for all "+
		"RangeSeries files for the site. If set, `siteDir/RangeSeries` will be scanned for RangeSeries files.")
	outputFileType := flag.String("output-file-type", "JSON", "The format of the output file. Options are 'JSON' or 'CSV'.")
	outputFileName := flag.String("output-file-name", "rangeseries_to_config", "The name of the output file. Should not include the file ending.")

	flag.Parse()

	targetRangeseriesFiles := flag.Args()

	fmt.Println("Target site directory:", *siteDir)
	fmt.Println("Output file type:", *outputFileType)
	fmt.Println("Output file name:", *outputFileName)
	fmt.Println("Targetting all RangeSeries files:", *allRangeSeries)

	return targetRangeseriesFiles, *allRangeSeries, *siteDir, *outputFileType, *outputFileName
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
	targetRangeseriesFiles, allRangeSeries, siteDir, outputFileType, outputFileName := parseArgs()
	validateArgs(targetRangeseriesFiles, allRangeSeries, siteDir, outputFileType, outputFileName)

	// 2. Build mapping of time intervals to configs
	// 2.a. Retrieve configs
	autoConfigs, _ := readConfigFiles(siteDir, "Config_Auto")
	operatorConfigs, _ := readConfigFiles(siteDir, "Config_Operator")

	// 2.b. Build mapping of time intervals to configs
	autoConfigIntervals := mapping.BuildAutoConfigIntervals(autoConfigs)
	operatorConfigIntervals := mapping.BuildOperatorConfigIntervals(operatorConfigs)

	// 3. Build mapping of RangeSeries files to Config directories
	var rangeSeriesFilePaths []string
	if allRangeSeries {
		rangeSeriesFilePaths, _ = readRangeseriesFiles(siteDir)
	} else {
		rangeSeriesFilePaths = targetRangeseriesFiles
	}

	rangeSeriesToConfig := mapping.CreateRangeSeriesToConfigMap(rangeSeriesFilePaths, autoConfigIntervals, operatorConfigIntervals)

	// 4. Write mapping to disk
	writeResult(rangeSeriesToConfig, outputFileType, outputFileName)

}
