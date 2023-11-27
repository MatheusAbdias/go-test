package util

import (
	"log"
	"os"
	"path/filepath"
)

var ProjectName = "web-test"

func findProjectDirRecursive(currentDir string) string {
	if base := filepath.Base(currentDir); base == ProjectName {
		return currentDir
	}

	if currentDir == "/" {
		return ""
	}

	parentDir := filepath.Dir(currentDir)
	return findProjectDirRecursive(parentDir)
}

func FindProjectDir() string {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return findProjectDirRecursive(currentDir)
}
