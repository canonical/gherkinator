// Package generate implements the `gherkinator generate` subcommand: it
// reads a YAML test plan, validates and optionally filters it, and
// transpiles the result into Gherkin (.feature) or Markdown (.md) files.
package generate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v28"
	messages "github.com/cucumber/messages/go/v24"
	"go.yaml.in/yaml/v3"

	"gherkinator/internal/common"
)

// ValidateGherkin parses a Gherkin string with the official Cucumber parser
// and returns an error if the syntax is invalid.
func ValidateGherkin(gherkinText string) error {
	reader := strings.NewReader(gherkinText)
	_, err := gherkin.ParseGherkinDocument(reader, (&messages.Incrementing{}).NewId)
	return err
}

// ProcessFile reads a YAML file (handling multi-document streams), validates
// schemas, transpiles to the requested format, and writes output files.
//
// riskFilter and statusFilter are intersected: a plan must satisfy both
// filters (or either filter, when its value is empty) to be rendered.
// Pass "" for either filter to disable that dimension of filtering.
func ProcessFile(filename string, format string, outputDir string, riskFilter string, statusFilter string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var plans []common.TestPlan
	decoder := yaml.NewDecoder(file)
	for i := 1; ; i++ {
		var plan common.TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode YAML document %d: %w", i, err)
		}

		if err := common.ValidateSchema(plan); err != nil {
			return fmt.Errorf("validation error in document %d: %w", i, err)
		}

		plans = append(plans, plan)
	}

	// Apply status filter first, then risk filter. Each filter is a no-op
	// when its argument is empty, so passing neither, one, or both filters
	// produces the expected intersection.
	filteredPlans := common.FilterPlansByStatus(plans, statusFilter)
	filteredPlans = common.FilterPlansByRisk(filteredPlans, riskFilter)

	var totalDocs int
	for i, plan := range filteredPlans {
		totalDocs = i + 1

		var output string
		var ext string
		switch format {
		case "gh":
			output = common.GenerateGherkin(plan)
			if err := ValidateGherkin(output); err != nil {
				return fmt.Errorf("generated Gherkin for document %d is invalid: %w", totalDocs, err)
			}
			ext = ".feature"
		case "md":
			output = common.GenerateMarkdown(plan)
			ext = ".md"
		default:
			return fmt.Errorf("unsupported format: %s", format)
		}

		safeFilename := strings.ReplaceAll(strings.ToLower(plan.Feature), " ", "_")
		if safeFilename == "" {
			safeFilename = fmt.Sprintf("plan_%d", totalDocs)
		}

		outPath := filepath.Join(outputDir, safeFilename+ext)
		if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file %s: %w", outPath, err)
		}
	}
	return nil
}
