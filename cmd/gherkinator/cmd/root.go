// Package cmd wires up the gherkinator cobra command tree. It contains
// one file per subcommand plus a shared root command.
package cmd

import (
	"github.com/canonical/gherkinator/internal/common"
	"github.com/canonical/gherkinator/internal/version"

	"github.com/spf13/cobra"
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

	// Wire the version string from the internal/version package into the
	// root command. We pre-register a boolean `--version` flag with no
	// `-v` shorthand so cobra's auto-version logic uses our definition
	// (cobra would otherwise add `-v` as a shorthand, but the issue for
	// this feature reserves `-v` for a future `--verbose` flag). The
	// custom template prints the bare version string rather than the
	// default "{Name} version {Version}" format.
	rootCmd.Flags().Bool("version", false, "Print version and exit")
	rootCmd.Version = version.Version
	rootCmd.SetVersionTemplate("{{.Version}}\n")

	rootCmd.AddCommand(initCmd, generateCmd, serveCmd, deleteCmd, cleanCmd, editCmd, validateCmd, versionCmd)
}

// Execute runs the gherkinator root command. It is called by main() and
// exits the process with a non-zero status on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		//nolint:gocritic // Exit code reflects Cobra's CLI contract
		_ = err
	}
}
