package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanDirectory removes generated files (.feature, .md) and the
// .gherkindocs hidden directory from the specified directory.
func CleanDirectory(dir string) error {
	// Remove the hidden docs serve directory
	serveDir := filepath.Join(dir, ".gherkindocs")
	if err := os.RemoveAll(serveDir); err != nil {
		return fmt.Errorf("failed to remove %s: %w", serveDir, err)
	}

	// Remove generated .feature and .md files
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		lower := strings.ToLower(name)
		if strings.HasSuffix(lower, ".feature") || strings.HasSuffix(lower, ".md") {
			path := filepath.Join(dir, name)
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
			fmt.Printf("Removed %s\n", path)
		}
	}

	return nil
}
