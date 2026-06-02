// Package serve implements the `gherkinator serve` subcommand: it generates
// Sphinx documentation from YAML test plans, watches for changes, and
// launches a local documentation server inside a Bubbletea TUI.
package serve

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"

	"gherkinator/internal/common"
)

// validTestTypes is the list of recognised test types used to create
// subdirectories inside the Sphinx docs directory.
var validTestTypes = []string{
	"functional",
	"solution",
	"performance",
	"reliability",
	"security",
}

// GenerateSphinxDocs reads a YAML test plan file and generates Markdown
// files organized by test type inside the docs directory.  Each plan is
// written to <docsDir>/<type>/<safe_feature_name>.md.  It returns the
// list of generated plans so callers can build a toctree.
//
// riskFilter and statusFilter are intersected: a plan must satisfy both
// filters (or either filter, when its value is empty) to be rendered.
// Pass "" for either filter to disable that dimension of filtering.
func GenerateSphinxDocs(yamlFile string, docsDir string, riskFilter string, statusFilter string) ([]common.TestPlan, error) {
	file, err := os.Open(yamlFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var plans []common.TestPlan
	decoder := yaml.NewDecoder(file)
	for i := 1; ; i++ {
		var plan common.TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML document %d: %w", i, err)
		}

		if err := common.ValidateSchema(plan); err != nil {
			return nil, fmt.Errorf("validation error in document %d: %w", i, err)
		}

		plans = append(plans, plan)
	}

	// Apply status filter first, then risk filter. Each filter is a no-op
	// when its argument is empty, so passing neither, one, or both filters
	// produces the expected intersection.
	filteredPlans := common.FilterPlansByStatus(plans, statusFilter)
	filteredPlans = common.FilterPlansByRisk(filteredPlans, riskFilter)

	for _, plan := range filteredPlans {
		output := common.GenerateMarkdown(plan)

		safeFilename := strings.ReplaceAll(strings.ToLower(plan.Feature), " ", "_")
		if safeFilename == "" {
			safeFilename = "plan"
		}

		typeDir := filepath.Join(docsDir, plan.Type)
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", typeDir, err)
		}

		outPath := filepath.Join(typeDir, safeFilename+".md")
		if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", outPath, err)
		}
	}

	return filteredPlans, nil
}

// BuildTypeLandingPages creates a sub-landing page for each test type
// that has plans.  Each landing page has a title and a toctree listing
// the generated plan files.
func BuildTypeLandingPages(docsDir string, grouped map[string][]string) error {
	for _, testType := range validTestTypes {
		files, ok := grouped[testType]
		if !ok || len(files) == 0 {
			continue
		}

		var builder strings.Builder
		//nolint:staticcheck // strings.Title is fine here
		fmt.Fprintf(&builder, "# %s\n\n", strings.Title(testType))
		builder.WriteString("```{toctree}\n")
		builder.WriteString(":maxdepth: 1\n\n")
		for _, f := range files {
			fmt.Fprintf(&builder, "%s\n", f)
		}
		builder.WriteString("```\n")

		typeDir := filepath.Join(docsDir, testType)
		if err := os.MkdirAll(typeDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", typeDir, err)
		}
		indexPath := filepath.Join(typeDir, "index.md")
		if err := os.WriteFile(indexPath, []byte(builder.String()), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", indexPath, err)
		}
	}
	return nil
}

// planEntry stores a feature name alongside its safe filename so that
// the root index can list human-readable bullet points.
type planEntry struct {
	Feature      string
	SafeFilename string
}

// BuildSphinxIndex writes a MyST-flavoured Markdown index.md with level-2
// headers for each test type and bullet-pointed feature names underneath.
// A hidden toctree is still emitted so that Sphinx includes the pages in
// its navigation tree.
func BuildSphinxIndex(docsDir string, plans []common.TestPlan) error {
	// Group plans by type (keep both display name and safe filename)
	grouped := make(map[string][]string)
	entries := make(map[string][]planEntry)
	for _, plan := range plans {
		safeFilename := strings.ReplaceAll(strings.ToLower(plan.Feature), " ", "_")
		if safeFilename == "" {
			safeFilename = "plan"
		}
		grouped[plan.Type] = append(grouped[plan.Type], safeFilename)
		entries[plan.Type] = append(entries[plan.Type], planEntry{
			Feature:      plan.Feature,
			SafeFilename: safeFilename,
		})
	}

	// Create sub-landing pages for each type
	if err := BuildTypeLandingPages(docsDir, grouped); err != nil {
		return fmt.Errorf("failed to build type landing pages: %w", err)
	}

	// Build root index with type headers and feature bullet points
	var builder strings.Builder
	builder.WriteString("# Test Plans\n\n")

	// Hidden toctree so Sphinx still discovers all pages
	builder.WriteString("```{toctree}\n")
	builder.WriteString(":hidden:\n")
	builder.WriteString(":maxdepth: 2\n\n")
	for _, testType := range validTestTypes {
		if _, ok := grouped[testType]; !ok || len(grouped[testType]) == 0 {
			continue
		}
		fmt.Fprintf(&builder, "%s/index\n", testType)
	}
	builder.WriteString("```\n\n")

	// Visible type sections with feature bullet points
	for _, testType := range validTestTypes {
		typeEntries, ok := entries[testType]
		if !ok || len(typeEntries) == 0 {
			continue
		}
		//nolint:staticcheck // strings.Title is fine here
		fmt.Fprintf(&builder, "## %s\n\n", strings.Title(testType))
		for _, entry := range typeEntries {
			displayName := entry.Feature
			if displayName == "" {
				displayName = entry.SafeFilename
			}
			fmt.Fprintf(&builder, "- [%s](%s/%s.md)\n", displayName, testType, entry.SafeFilename)
		}
		builder.WriteString("\n")
	}

	indexPath := filepath.Join(docsDir, "index.md")
	if err := os.WriteFile(indexPath, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", indexPath, err)
	}
	return nil
}

// UpdateConfPy rewrites the Sphinx conf.py file to:
//   - set the `project` variable to the given name
//   - uncomment `disable_feedback_button = True`
//   - remove rediraffe-related configuration lines
//   - remove "sphinx_rerediraffe" from the extensions list
func UpdateConfPy(confPath string, projectName string) error {
	data, err := os.ReadFile(confPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", confPath, err)
	}

	lines := strings.Split(string(data), "\n")
	var out []string
	foundProject := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Set project name
		if !foundProject && strings.HasPrefix(trimmed, "project") && strings.Contains(trimmed, "=") {
			out = append(out, fmt.Sprintf("project = %q", projectName))
			foundProject = true
			continue
		}

		// Uncomment disable_feedback_button
		if trimmed == "# disable_feedback_button = True" {
			out = append(out, "disable_feedback_button = True")
			continue
		}

		// Remove rediraffe configuration lines
		if strings.HasPrefix(trimmed, "rediraffe_redirects") ||
			strings.HasPrefix(trimmed, "rediraffe_dir_only") {
			continue
		}

		// Remove sphinx_rerediraffe from extensions list
		if strings.Contains(trimmed, `"sphinx_rerediraffe"`) {
			continue
		}

		out = append(out, line)
	}

	if !foundProject {
		return fmt.Errorf("could not find 'project' variable in %s", confPath)
	}

	if err := os.WriteFile(confPath, []byte(strings.Join(out, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", confPath, err)
	}
	return nil
}

// PrepareSphinxSite is the high-level orchestrator called by the serve
// command.  It operates on the docs/ subdirectory of the cloned slim
// starter pack: generates docs into type subdirectories, builds the
// toctree index with sub-landing pages per test type, and sets the
// project name in conf.py.
//
// riskFilter and statusFilter are intersected (see GenerateSphinxDocs).
func PrepareSphinxSite(yamlFile string, cloneDir string, projectName string, riskFilter string, statusFilter string) error {
	docsDir := filepath.Join(cloneDir, "docs")

	plans, err := GenerateSphinxDocs(yamlFile, docsDir, riskFilter, statusFilter)
	if err != nil {
		return fmt.Errorf("failed to generate sphinx docs: %w", err)
	}

	if err := BuildSphinxIndex(docsDir, plans); err != nil {
		return fmt.Errorf("failed to build sphinx index: %w", err)
	}

	confPath := filepath.Join(docsDir, "conf.py")
	if err := UpdateConfPy(confPath, projectName); err != nil {
		return fmt.Errorf("failed to update conf.py: %w", err)
	}

	return nil
}

// CleanGeneratedDocs removes the per-test-type subdirectories inside the
// Sphinx docs directory so that a subsequent regeneration does not leave
// stale markdown files for features that have been removed.  Files in the
// docs root (e.g. conf.py, Makefile, index.md) are preserved.
func CleanGeneratedDocs(docsDir string) error {
	for _, testType := range validTestTypes {
		typeDir := filepath.Join(docsDir, testType)
		if err := os.RemoveAll(typeDir); err != nil {
			return fmt.Errorf("failed to remove %s: %w", typeDir, err)
		}
	}
	return nil
}
