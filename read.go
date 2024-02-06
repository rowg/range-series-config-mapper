package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func FindFilesMatchingPattern(baseDir string, pattern string, wantDirectories bool) ([]string, error) {
	var matchingFiles []string
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				fmt.Printf("Warning: Permission denied accessing %s, skipping.\n", path)
				return nil
			}

			fmt.Println("Unhandled error while finding files:", err)
			return err
		}

		// Check if the file is the correct type (file/directory) and if it matches the pattern
		if d.IsDir() == wantDirectories && re.MatchString(path) {
			matchingFiles = append(matchingFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return matchingFiles, nil
}
