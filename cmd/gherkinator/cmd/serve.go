package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gherkinator/internal/common"
	"gherkinator/internal/serve"
)

var serveName string

// serveCmd generates Sphinx documentation, sets up file-watching for
// live reload, and launches `make run` inside a Bubbletea TUI.
//
// The --risk and --status flags bind to the shared riskFilter and
// statusFilter package-level variables (see generate.go).
var serveCmd = &cobra.Command{
	Use:   "serve [files-or-directories...]",
	Short: "Serve test plan docs and watch for changes",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if riskFilter != "" && !common.IsValidRisk(riskFilter) {
			return fmt.Errorf("--risk must be one of 'edge', 'beta', 'candidate', or 'stable'")
		}
		if statusFilter != "" && !common.IsValidStatus(statusFilter) {
			return fmt.Errorf("--status must be one of 'planned', 'implemented', or 'deprecated'")
		}
		inputFiles, err := common.DiscoverYAMLFiles(args)
		if err != nil {
			return fmt.Errorf("failed to resolve input files: %w", err)
		}

		return serve.Run(inputFiles, serveName, riskFilter, statusFilter)
	},
}

func init() {
	serveCmd.Flags().StringVarP(&serveName, "name", "n", "", "Project name for the documentation (defaults to current working directory name)")
	serveCmd.Flags().StringVar(&riskFilter, "risk", "", "Filter by risk level (edge, beta, candidate, stable)")
	serveCmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status (planned, implemented, deprecated)")
}
