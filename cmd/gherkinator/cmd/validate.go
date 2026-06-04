package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"gherkinator/internal/common"
)

var validateCmd = &cobra.Command{
	Use:   "validate [files-or-directories...]",
	Short: "Validate test plan YAML files",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		inputFiles, err := common.DiscoverYAMLFiles(args)
		if err != nil {
			return fmt.Errorf("failed to resolve input files: %w", err)
		}

		var errs []error
		totalPlans := 0
		for _, file := range inputFiles {
			plans, err := common.LoadTestPlans(file)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to read '%s': %w", file, err))
				continue
			}
			for i, plan := range plans {
				if err := common.ValidateSchema(plan); err != nil {
					errs = append(errs, fmt.Errorf(
						"test plan '%s' failed validation (document %d): %w",
						file, i+1, err,
					))
				}
			}
			totalPlans += len(plans)
		}

		if errs != nil {
			return errors.Join(errs...)
		}
		fmt.Printf("All %d test plan(s) are valid.\n", totalPlans)
		return nil
	},
}
