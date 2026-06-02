package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	gherkin "github.com/cucumber/gherkin/go/v28"
	messages "github.com/cucumber/messages/go/v24"
	"go.yaml.in/yaml/v3"
)

var riskOrder = map[string]int{
	"edge":      0,
	"beta":      1,
	"candidate": 2,
	"stable":    3,
}

func filterPlansByRisk(plans []TestPlan, filterRisk string) []TestPlan {
	if filterRisk == "" {
		return plans
	}
	maxLevel, ok := riskOrder[filterRisk]
	if !ok {
		return nil
	}
	var filtered []TestPlan
	for _, plan := range plans {
		level, ok := riskOrder[plan.Risk]
		if ok && level <= maxLevel {
			filtered = append(filtered, plan)
		}
	}
	return filtered
}

// ValidateGherkin parses a Gherkin string with the official Cucumber parser
// and returns an error if the syntax is invalid.
func ValidateGherkin(gherkinText string) error {
	reader := strings.NewReader(gherkinText)
	_, err := gherkin.ParseGherkinDocument(reader, (&messages.Incrementing{}).NewId)
	return err
}

// ProcessFile reads a YAML file (handling multi-document streams), validates
// schemas, transpiles to the requested format, and writes output files.
func ProcessFile(filename string, format string, outputDir string, riskFilter string) error {
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

	var plans []TestPlan
	decoder := yaml.NewDecoder(file)
	for i := 1; ; i++ {
		var plan TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode YAML document %d: %w", i, err)
		}

		if err := ValidateSchema(plan); err != nil {
			return fmt.Errorf("validation error in document %d: %w", i, err)
		}

		plans = append(plans, plan)
	}

	filteredPlans := filterPlansByRisk(plans, riskFilter)

	var totalDocs int
	for i, plan := range filteredPlans {
		totalDocs = i + 1

		var output string
		var ext string
		switch format {
		case "gh":
			output = GenerateGherkin(plan)
			if err := ValidateGherkin(output); err != nil {
				return fmt.Errorf("generated Gherkin for document %d is invalid: %w", totalDocs, err)
			}
			ext = ".feature"
		case "md":
			output = GenerateMarkdown(plan)
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
