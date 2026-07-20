package common

import (
	"fmt"
	"io"
	"os"

	"go.yaml.in/yaml/v3"
)

// RenderedFile is an in-memory representation of a generated feature
// file produced by RenderPlans.  It pairs the on-disk filename (e.g.
// "submit_mpi_job.feature") with the rendered text content and the
// original plan's display name so callers can preserve case in
// human-readable output.
type RenderedFile struct {
	// Filename is the on-disk filename (including extension) the
	// rendered Content would be written to.
	Filename string
	// Content is the rendered text (Gherkin or Markdown).
	Content string
	// FeatureName is the original case-preserving title from the
	// source TestPlan.  Used by callers that need to display
	// human-readable plan names in their output.
	FeatureName string
}

// RenderPlans loads YAML test plans from the given input files,
// validates each plan's schema and (for gh format) the generated
// Gherkin output, applies the optional risk/status filters, and
// returns each surviving plan as a RenderedFile.  No disk I/O is
// performed: callers handle writing or comparing the renderings.
//
// The format argument must be "gh" (Gherkin, ".feature" output) or
// "md" (Markdown, ".md" output).
//
// riskFilter and statusFilter are intersected: a plan must satisfy
// both filters (or either filter, when its value is empty) to be
// rendered.  Pass "" for either filter to disable that dimension.
//
// Plans with an empty feature field are rendered using a stable
// fallback basename ("plan_N") so that empty names produce a
// deterministic, non-empty filename regardless of format.
func RenderPlans(inputFiles []string, format string, riskFilter string, statusFilter string) ([]RenderedFile, error) {
	if format != "gh" && format != "md" {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	var plans []TestPlan
	for _, filename := range inputFiles {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
		}

		docPlans, err := decodeAndValidatePlans(file)
		closeErr := file.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, fmt.Errorf("failed to close file %s: %w", filename, closeErr)
		}
		plans = append(plans, docPlans...)
	}

	filteredPlans := FilterPlansByStatus(plans, statusFilter)
	filteredPlans = FilterPlansByRisk(filteredPlans, riskFilter)

	var rendered []RenderedFile
	for i, plan := range filteredPlans {
		docNum := i + 1
		var content string
		switch format {
		case "gh":
			content = GenerateGherkin(plan)
			if err := ValidateGherkin(content); err != nil {
				return nil, fmt.Errorf("document %d: generated Gherkin is invalid: %w", docNum, err)
			}
		case "md":
			content = GenerateMarkdown(plan)
		}

		fallback := fallbackName(format, docNum)
		filename := SafeFeatureName(plan, fallback) + extension(format)
		rendered = append(rendered, RenderedFile{
			Filename:    filename,
			Content:     content,
			FeatureName: plan.Feature,
		})
	}
	return rendered, nil
}

// decodeAndValidatePlans decodes a multi-document YAML stream from r
// and validates each document's schema.  On the first validation or
// decode error the document's 1-based index is included in the error
// message so the caller can identify which entry is bad.
func decodeAndValidatePlans(r io.Reader) ([]TestPlan, error) {
	var plans []TestPlan
	decoder := yaml.NewDecoder(r)
	for i := 1; ; i++ {
		var plan TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML document %d: %w", i, err)
		}
		if err := ValidateSchema(plan); err != nil {
			return nil, fmt.Errorf("validation error in document %d: %w", i, err)
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// extension returns the on-disk file extension used by a given format.
func extension(format string) string {
	switch format {
	case "gh":
		return ".feature"
	case "md":
		return ".md"
	default:
		return ""
	}
}

// fallbackName returns the deterministic basename to use when a plan's
// feature field is empty.  The 1-based document index is always included
// so multi-document YAMLs produce distinct filenames and the behaviour
// is identical across gh and md formats.
func fallbackName(format string, docNum int) string {
	return fmt.Sprintf("plan_%d", docNum)
}
