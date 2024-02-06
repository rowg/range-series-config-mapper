package write

import (
	"encoding/csv"
	"encoding/json"
	"log"
	"os"
)

const jsonFileEnding = ".json"
const csvFileEnding = ".csv"

func SaveMapAsJson(myMap map[string]string, fileName string) {
	jsonData, err := json.Marshal(myMap)
	if err != nil {
		log.Fatalf("Error marshalling map to JSON: %v", err)
	}

	// Write JSON data to file map.json in current directory
	err = os.WriteFile(fileName+jsonFileEnding, jsonData, 0644)
	if err != nil {
		log.Fatalf("Error writing JSON to file: %v", err)
	}
}

func SaveMapAsCsv(myMap map[string]string, fileName string) {
	// Create a new CSV file
	file, err := os.Create(fileName + csvFileEnding)
	if err != nil {
		log.Fatalf("Error creating CSV file: %v", err)
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Iterate over the map and write each key-value pair as a row in the CSV file
	for key, value := range myMap {
		err := writer.Write([]string{key, value})
		if err != nil {
			log.Fatalf("Error writing to CSV file: %v", err)
		}
	}
}
