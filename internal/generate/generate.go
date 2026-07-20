// Package generate implements the `gherkinator generate` subcommand: it
// reads a YAML test plan, validates and optionally filters it, and
// transpiles the result into Gherkin (.feature) or Markdown (.md) files
// on disk.
package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/gherkinator/internal/common"
)

// ProcessFile reads a YAML file (handling multi-document streams), validates
// schemas, transpiles to the requested format, and writes output files.
//
// riskFilter and statusFilter are intersected: a plan must satisfy both
// filters (or either filter, when its value is empty) to be rendered.
// Pass "" for either filter to disable that dimension of filtering.
//
// Internally this delegates the load/validate/filter/render pipeline to
// common.RenderPlans and only owns the disk-write concern specific to
// `gherkinator generate`: writing files flat into outputDir.
func ProcessFile(filename string, format string, outputDir string, riskFilter string, statusFilter string) error {
	rendered, err := common.RenderPlans([]string{filename}, format, riskFilter, statusFilter)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, rf := range rendered {
		outPath := filepath.Join(outputDir, rf.Filename)
		if err := os.WriteFile(outPath, []byte(rf.Content), 0644); err != nil {
			return fmt.Errorf("failed to write output file %s: %w", outPath, err)
		}
	}
	return nil
}
