package cmd

import (
	"github.com/spf13/cobra"

	"github.com/canonical/gherkinator/internal/version"
)

// versionCmd prints the gherkinator version string. The same string is
// also reachable via the root command's --version flag.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the gherkinator version",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Println(version.Version)
		return nil
	},
}
