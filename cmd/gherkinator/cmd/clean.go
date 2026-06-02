package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"gherkinator/internal/clean"
)

var cleanDir string

// cleanCmd removes generated files (.feature, .md) and the .gherkindocs/
// staging directory from the target directory.
var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean generated files and temporary directories from the test plan directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := clean.CleanDirectory(cleanDir); err != nil {
			return fmt.Errorf("clean failed: %w", err)
		}
		fmt.Println("Clean complete.")
		return nil
	},
}

func init() {
	cleanCmd.Flags().StringVarP(&cleanDir, "dir", "d", ".", "Directory to clean")
}
