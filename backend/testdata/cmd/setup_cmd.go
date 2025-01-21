package main

import (
	"log"
	"os"
	"path/filepath"

	"sample-exchange/backend/config"
	"sample-exchange/backend/testdata"
)

func main() {
	// Load config from environment
	cfg := &config.Config{
		DevMode:           true,
		BypassTimeWindows: true,
		BypassOAuth:       true,
	}

	// Ensure we're in project root
	if err := os.Chdir(findProjectRoot()); err != nil {
		log.Fatalf("Failed to change to project root: %v", err)
	}

	testData, err := testdata.Setup(cfg)
	if err != nil {
		log.Fatalf("Failed to setup test data: %v", err)
	}

	if err := testData.UpdateBrunoEnv(); err != nil {
		log.Printf("Warning: Failed to update Bruno env: %v", err)
	}

	log.Printf("Setup complete with token: %s", testData.Token)
}

// findProjectRoot walks up directories until it finds go.mod
func findProjectRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "."
		}
		dir = parent
	}
}