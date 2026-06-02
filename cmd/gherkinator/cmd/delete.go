package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gherkinator/internal/common"
	"gherkinator/internal/delete"
)

var (
	skipConfirm     bool
	deleteInputFile string
)

// deleteCmd removes test plan entries from a YAML file by feature name.
var deleteCmd = &cobra.Command{
	Use:   "delete [feature-names...]",
	Short: "Delete test plans by feature name from a YAML file",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if deleteInputFile == "" {
			return fmt.Errorf("--input/-i is required")
		}
		plans, err := common.LoadTestPlans(deleteInputFile)
		if err != nil {
			return fmt.Errorf("failed to load test plans: %w", err)
		}

		remaining, deleted := delete.DeleteTestPlans(plans, args)
		if len(deleted) == 0 {
			fmt.Println("No matching test plans found.")
			return nil
		}

		if !skipConfirm {
			if !delete.ConfirmDeletion(deleted, os.Stdin) {
				fmt.Println("Delete aborted.")
				return nil
			}
		}

		if err := common.WriteTestPlans(deleteInputFile, remaining); err != nil {
			return fmt.Errorf("failed to write updated test plans: %w", err)
		}

		fmt.Printf("Successfully deleted %d test plan(s).\n", len(deleted))
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")
	deleteCmd.Flags().StringVarP(&deleteInputFile, "input", "i", "", "Path to the input YAML file (required)")
}
