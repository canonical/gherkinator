// Package diff implements the `gherkinator diff` subcommand: it
// compares a set of YAML test plans against an on-disk features
// directory and reports missing, differing, and orphan feature files.
package diff

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"znkr.io/diff/textdiff"

	"github.com/canonical/gherkinator/internal/common"
)

// Exit codes returned by Run.  They are re-exported so the cmd package
// can propagate them via os.Exit.
const (
	ExitClean     = 0
	ExitDiff      = 1
	ExitPlanError = 2
)

// Run compares the rendered plans (built from inputFiles and filtered
// by opts) against the .feature files in featuresDir, writing any
// human-readable classifications to errOut and (when show is true)
// unified diffs for differing pairs to out.
//
// It returns the appropriate exit code: ExitClean on no diffs,
// ExitDiff when any classification fires, or ExitPlanError when a
// plan could not be loaded, validated, or rendered.
func Run(inputFiles []string, featuresDir string, opts common.FilterOptions, show bool, out, errOut io.Writer) int {
	rendered, err := common.RenderPlans(inputFiles, "gh", opts)
	if err != nil {
		//nolint:errcheck // Writing to stderr; error is not actionable
		fmt.Fprintf(errOut, "Error: %s\n", err)
		return ExitPlanError
	}

	expected, duplicate := indexRendered(rendered)
	if duplicate != "" {
		//nolint:errcheck // Writing to stderr; error is not actionable
		fmt.Fprintf(errOut, "Error: two or more plans produce the same feature file \"%s\".\n", duplicate)
		return ExitPlanError
	}

	actual, err := scanFeatureDir(featuresDir)
	if err != nil {
		//nolint:errcheck // Writing to stderr; error is not actionable
		fmt.Fprintf(errOut, "Error: %s\n", err)
		return ExitPlanError
	}

	missing := findMissing(expected, actual)
	differing := findDiffering(expected, actual)
	orphans := findOrphans(expected, actual)

	reportMissing(missing, errOut)
	reportOrphans(orphans, errOut)
	reportDiffering(differing, errOut)

	if show {
		for _, d := range differing {
			if err := writeUnifiedDiff(out, featuresDir, d); err != nil {
				//nolint:errcheck // Writing to stderr; error is not actionable
				fmt.Fprintf(errOut, "Error: failed to render diff for %s: %s\n", d.Filename, err)
				return ExitPlanError
			}
		}
	}

	if len(missing) > 0 || len(differing) > 0 || len(orphans) > 0 {
		return ExitDiff
	}
	return ExitClean
}

// classification represents a single planned feature file (by Filename)
// paired with its in-memory rendered Content.  It is the value type of
// the expected map; "missing" classifications are keys present in
// expected but absent in actual, "orphans" are the inverse.
//
// FeatureName preserves the original case-preserving title from the
// source plan so user-facing output doesn't have to lossy-decode the
// lowercased, underscore-spaced on-disk filename.
type classification struct {
	Filename    string
	Content     string
	FeatureName string
}

// indexRendered converts a slice of rendered files into a map keyed by
// filename.  If two rendered files share a filename, the duplicated
// filename is returned as the second value (empty when unique); the
// caller is responsible for formatting an error message and exiting.
func indexRendered(rendered []common.RenderedFile) (map[string]classification, string) {
	indexed := make(map[string]classification, len(rendered))
	for _, rf := range rendered {
		if _, exists := indexed[rf.Filename]; exists {
			return nil, rf.Filename
		}
		indexed[rf.Filename] = classification{
			Filename:    rf.Filename,
			Content:     rf.Content,
			FeatureName: rf.FeatureName,
		}
	}
	return indexed, ""
}

// scanFeatureDir returns a map of every .feature file in featuresDir
// (non-recursive) keyed by filename, with the file's contents as the
// value.  If featuresDir does not exist or is not a directory, an
// error is returned (treated as a plan error by Run).
func scanFeatureDir(featuresDir string) (map[string]string, error) {
	info, err := os.Stat(featuresDir)
	if err != nil {
		return nil, fmt.Errorf("features directory %q: %w", featuresDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("features path %q is not a directory", featuresDir)
	}

	entries, err := os.ReadDir(featuresDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read features directory %q: %w", featuresDir, err)
	}

	files := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(e.Name()), ".feature") {
			continue
		}
		path := filepath.Join(featuresDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read feature file %s: %w", path, err)
		}
		files[e.Name()] = string(data)
	}
	return files, nil
}

// findMissing returns the entries in expected whose filename is not
// present in actual.  The returned slice is sorted by filename for
// deterministic output.
func findMissing(expected map[string]classification, actual map[string]string) []classification {
	var missing []classification
	for name, exp := range expected {
		if _, ok := actual[name]; !ok {
			missing = append(missing, exp)
		}
	}
	sort.Slice(missing, func(i, j int) bool { return missing[i].Filename < missing[j].Filename })
	return missing
}

// findOrphans returns the filenames in actual that have no matching
// entry in expected, sorted for deterministic output.
func findOrphans(expected map[string]classification, actual map[string]string) []string {
	var orphans []string
	for name := range actual {
		if _, ok := expected[name]; !ok {
			orphans = append(orphans, name)
		}
	}
	sort.Strings(orphans)
	return orphans
}

// findDiffering returns the entries whose plan-rendered content does
// not match the on-disk content.  Sorted for deterministic output.
//
// This function is named to reflect that it is the inverse of the
// "missing" and "orphan" classifications: a feature file that exists
// and maps to a plan but whose contents do not match.
func findDiffering(expected map[string]classification, actual map[string]string) []classification {
	var differing []classification
	for name, exp := range expected {
		got, ok := actual[name]
		if !ok {
			continue
		}
		if got != exp.Content {
			differing = append(differing, exp)
		}
	}
	sort.Slice(differing, func(i, j int) bool { return differing[i].Filename < differing[j].Filename })
	return differing
}

// reportMissing writes one line summarising all missing classifications
// to w.  The verb agrees with the count; the format follows the
// example in issue #3.
func reportMissing(missing []classification, w io.Writer) {
	if len(missing) == 0 {
		return
	}
	names := make([]string, len(missing))
	for i, m := range missing {
		names[i] = m.FeatureName
	}
	verb := "is"
	if len(missing) > 1 {
		verb = "are"
	}
	//nolint:errcheck // Writing to stderr; error is not actionable
	fmt.Fprintf(w, "Feature file for plan(s) %s %s missing.\n", quotedList(names), verb)
}

// reportOrphans writes one line summarising all orphan feature file
// names.  Format follows the example in issue #3.
func reportOrphans(orphans []string, w io.Writer) {
	if len(orphans) == 0 {
		return
	}
	verb := "matches"
	if len(orphans) > 1 {
		verb = "match"
	}
	//nolint:errcheck // Writing to stderr; error is not actionable
	fmt.Fprintf(w, "Feature file(s) %s %s no test plan(s).\n", quotedList(orphans), verb)
}

// reportDiffering writes one line per group of differing pairs grouped
// on a single line, matching the format shown in issue #3's example.
func reportDiffering(differing []classification, w io.Writer) {
	if len(differing) == 0 {
		return
	}
	plans := make([]string, len(differing))
	files := make([]string, len(differing))
	for i, d := range differing {
		plans[i] = d.FeatureName
		files[i] = d.Filename
	}
	verb := "differs"
	if len(differing) > 1 {
		verb = "differ"
	}
	//nolint:errcheck // Writing to stderr; error is not actionable
	fmt.Fprintf(w, "Plan(s) %s %s from feature file(s) %s.\n", quotedList(plans), verb, quotedList(files))
}

// writeUnifiedDiff prints a unified diff between the generated plan
// content and the on-disk feature file in the format produced by the
// `diff -u` command: two filename headers followed by the hunks.  The
// znkr/diff library omits the `---`/`+++` headers itself, so we
// prepend them to match.
func writeUnifiedDiff(w io.Writer, featuresDir string, c classification) error {
	actual, err := os.ReadFile(filepath.Join(featuresDir, c.Filename))
	if err != nil {
		return fmt.Errorf("read on-disk file: %w", err)
	}

	generatedPath := fmt.Sprintf("a/%s", c.Filename)
	onDiskPath := fmt.Sprintf("b/%s", c.Filename)
	if _, err := fmt.Fprintf(w, "--- %s\n", generatedPath); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "+++ %s\n", onDiskPath); err != nil {
		return err
	}

	body := textdiff.Unified(c.Content, string(actual))
	if _, err := io.WriteString(w, body); err != nil {
		return err
	}
	return nil
}

// quotedList joins a slice of strings with ", " between each, wrapping
// each value in double quotes.  Empty inputs return an empty string.
func quotedList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	quoted := make([]string, len(items))
	for i, s := range items {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(quoted, ", ")
}
