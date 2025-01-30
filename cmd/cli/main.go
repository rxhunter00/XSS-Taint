package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rxhunter00/XSS-Taint/pkg/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: [directory path] [optional output path]")
		os.Exit(1)
	}

	srcPath := os.Args[1]
	outPath := getOutputPath(srcPath)

	start := time.Now()

	filePaths, err := getPhpFiles(srcPath)
	if err != nil {
		log.Fatalf("Error getting PHP files: %v", err)
	}
	fmt.Printf("Scanning %d PHP files...\n", len(filePaths))

	result := scanner.Scan(srcPath, filePaths)

	elapsed := time.Since(start)
	fmt.Printf("Detected %d XSS vulnerabilities in %.2f seconds.\n", result.TotalFinding, elapsed.Seconds())

	if err := saveJSON(outPath, result); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}
}

// getOutputPath determines the output file path based on input arguments.
func getOutputPath(srcPath string) string {
	folderName := filepath.Base(srcPath)
	outPath := "results-" + folderName + ".json"

	if len(os.Args) > 2 {
		outPath = os.Args[2]
	}
	return outPath
}

// getPhpFiles recursively scans the directory and returns a list of PHP files.
func getPhpFiles(dirPath string) ([]string, error) {
	var files []string
	extension := ".php"

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() == "vendor" {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), extension) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// saveJSON writes the scan result to a JSON file.
func saveJSON(outPath string, result interface{}) error {
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()

	jsonEncoder := json.NewEncoder(f)
	jsonEncoder.SetEscapeHTML(false)
	return jsonEncoder.Encode(result)
}
