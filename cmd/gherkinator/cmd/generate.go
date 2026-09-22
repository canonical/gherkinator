package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/canonical/gherkinator/internal/common"
	"github.com/canonical/gherkinator/internal/generate"
)

// riskFilter, statusFilter, and tagFilters are shared package-level
// variables that the --risk, --status, and --tag flags of the generate,
// serve, and diff subcommands bind to.  Sharing them ensures that
// running one subcommand followed by the other in the same process
// (e.g. during tests) leaves no stale state behind.
var (
	outputDir    string
	format       string
	riskFilter   string
	statusFilter string
	tagFilters   []string
)

// validateTagFilters reports an error if any value bound to the shared
// --tag flag fails the tag format rules (see common.ValidateTag).
func validateTagFilters() error {
	for _, tag := range tagFilters {
		if err := common.ValidateTag(tag); err != nil {
			return fmt.Errorf("--tag: %w", err)
		}
	}
	return nil
}

// generateCmd generates Gherkin (.feature) or Markdown (.md) files from
// YAML test plan inputs.
var generateCmd = &cobra.Command{
	Use:   "generate [files-or-directories...]",
	Short: "Generate gherkin (gh) or markdown (md) files",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if format != "gh" && format != "md" {
			return fmt.Errorf("--format must be either 'gh' or 'md'")
		}
		if riskFilter != "" && !common.IsValidRisk(riskFilter) {
			return fmt.Errorf("--risk must be one of 'edge', 'beta', 'candidate', or 'stable'")
		}
		if statusFilter != "" && !common.IsValidStatus(statusFilter) {
			return fmt.Errorf("--status must be one of 'planned', 'implemented', or 'deprecated'")
		}
		if err := validateTagFilters(); err != nil {
			return err
		}
		filterOpts := common.FilterOptions{Risk: riskFilter, Status: statusFilter, Tags: tagFilters}
		inputFiles, err := common.DiscoverYAMLFiles(args)
		if err != nil {
			return fmt.Errorf("failed to resolve input files: %w", err)
		}
		var errs []error
		for _, file := range inputFiles {
			if err := generate.ProcessFile(file, format, outputDir, filterOpts); err != nil {
				errs = append(errs, fmt.Errorf("generation failed for %s: %w", file, err))
			}
		}
		if errs != nil {
			return errors.Join(errs...)
		}
		fmt.Println("Generation complete.")
		return nil
	},
}

func init() {
	generateCmd.Flags().StringVarP(&outputDir, "output-dir", "o", ".", "Directory to save generated files")
	generateCmd.Flags().StringVar(&format, "format", "gh", "Output format (gh or md)")
	generateCmd.Flags().StringVar(&riskFilter, "risk", "", "Filter by risk level (edge, beta, candidate, stable)")
	generateCmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status (planned, implemented, deprecated)")
	generateCmd.Flags().StringSliceVar(&tagFilters, "tag", nil,
		"Filter by tag; repeatable (union) and combinable with --risk and --status")
}
