package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gherkinator/internal/common"
	"gherkinator/internal/serve"
)

var (
	serveName string
	serveRisk string
)

// serveCmd generates Sphinx documentation, sets up file-watching for
// live reload, and launches `make run` inside a Bubbletea TUI.
var serveCmd = &cobra.Command{
	Use:   "serve [files-or-directories...]",
	Short: "Serve test plan docs and watch for changes",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if serveRisk != "" && !common.IsValidRisk(serveRisk) {
			return fmt.Errorf("--risk must be one of 'edge', 'beta', 'candidate', or 'stable'")
		}
		inputFiles, err := common.DiscoverYAMLFiles(args)
		if err != nil {
			return fmt.Errorf("failed to resolve input files: %w", err)
		}

		return serve.Run(inputFiles, serveName, serveRisk)
	},
}

func init() {
	serveCmd.Flags().StringVarP(&serveName, "name", "n", "", "Project name for the documentation (defaults to current working directory name)")
	serveCmd.Flags().StringVar(&serveRisk, "risk", "", "Filter by risk level (edge, beta, candidate, stable)")
}
