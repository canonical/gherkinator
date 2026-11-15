// Package clean implements the `gherkinator clean` subcommand: it removes
// generated documentation files and the serve staging directory.
package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/canonical/gherkinator/internal/common"
)

// CleanDirectory removes generated files (.feature, .md) and the serve
// staging directory (see common.StagingDirFor) from the specified directory.
func CleanDirectory(dir string) error {
	serveDir, err := common.StagingDirFor(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve serve directory: %w", err)
	}

	// Failures to remove the staging directory are warnings: removing the
	// generated files below is the primary purpose of this command.
	if err := os.RemoveAll(serveDir); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not fully remove serve directory %s: %v\n", serveDir, err) //nolint:errcheck // Writing to stderr; error is not actionable
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
