package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/canonical/gherkinator/internal/common"
	"github.com/canonical/gherkinator/internal/diff"
)

// showDiff is the package-level variable bound to the --show flag on
// the `gherkinator diff` subcommand.  When true, Run prints a unified
// diff to stdout for every classification whose on-disk feature file
// differs from the rendered plan.
var showDiff bool

// diffCmd evaluates whether the YAML test plans in <plans> differ
// from the .feature files generated under <features>.
//
// The <plans> argument follows common.DiscoverYAMLFiles semantics and
// may be a file path or directory; <features> must be a directory
// containing the on-disk feature files to compare against.
//
// The --risk, --status, and --tag flags are interpreted identically to
// the `generate` and `serve` subcommands and restrict which plans are
// considered during the comparison.
//
// Exit codes:
//
//	0  no diffs
//	1  at least one diff classification was reported
//	2  plan or features-directory error
//
// cobra's RunE returns an error for unexpected CLI misuse (bad flags,
// argument validation, etc.); process-level exit is performed
// directly via os.Exit because cobra only exposes a single non-zero
// status on error and the diff command needs distinct codes for
// "diffs found" and "plan error".
var diffCmd = &cobra.Command{
	Use:   "diff [flags] <plans> <features>",
	Short: "Diff gherkinator test plans against generated feature files",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if riskFilter != "" && !common.IsValidRisk(riskFilter) {
			return fmt.Errorf("--risk must be one of 'edge', 'beta', 'candidate', or 'stable'")
		}
		if statusFilter != "" && !common.IsValidStatus(statusFilter) {
			return fmt.Errorf("--status must be one of 'planned', 'implemented', or 'deprecated'")
		}
		if err := validateTagFilters(); err != nil {
			return err
		}

		inputFiles, err := common.DiscoverYAMLFiles([]string{args[0]})
		if err != nil {
			return fmt.Errorf("failed to resolve plan inputs: %w", err)
		}

		filterOpts := common.FilterOptions{Risk: riskFilter, Status: statusFilter, Tags: tagFilters}
		os.Exit(diff.Run(inputFiles, args[1], filterOpts, showDiff, os.Stdout, os.Stderr))
		return nil
	},
}

func init() {
	diffCmd.Flags().StringVar(&riskFilter, "risk", "", "Filter by risk level (edge, beta, candidate, stable)")
	diffCmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status (planned, implemented, deprecated)")
	diffCmd.Flags().StringSliceVar(&tagFilters, "tag", nil,
		"Filter by tag; repeatable (union) and combinable with --risk and --status")
	diffCmd.Flags().BoolVar(&showDiff, "show", false, "Print unified diffs for differing feature files to stdout")
}
