// Package cmd wires up the gherkinator cobra command tree. It contains
// one file per subcommand plus a shared root command.
package cmd

import (
	"github.com/spf13/cobra"

	"gherkinator/internal/common"
)

var rootCmd = &cobra.Command{
	Use:   "gherkinator",
	Short: "A testing plan management and generation tool",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// init registers all subcommands on rootCmd.
func init() {
	// Initialise Viper (defaults, config file lookup, env overrides) before
	// any subcommand runs.
	cobra.OnInitialize(common.InitConfig)
	rootCmd.AddCommand(initCmd, generateCmd, serveCmd, deleteCmd, cleanCmd, editCmd, validateCmd)
}

// Execute runs the gherkinator root command. It is called by main() and
// exits the process with a non-zero status on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//nolint:gocritic // Exit code reflects Cobra's CLI contract
		_ = err
	}
}
