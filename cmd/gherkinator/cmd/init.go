package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"gherkinator/internal/common"
)

// planName is the name of the YAML file created by the init command.
// The .yaml extension is appended automatically if not already present.
var planName string

// initCmd creates a new directory containing an empty test plan template.
var initCmd = &cobra.Command{
	Use:   "init [directory-name]",
	Short: "Initialize a new test plan directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirName := args[0]
		if err := os.MkdirAll(dirName, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		if planName == "" {
			return fmt.Errorf("--name cannot be empty")
		}
		fileName := planName
		if !common.HasYAMLExt(fileName) {
			fileName += ".yaml"
		}

		emptyPlan := `---
feature: ""
type: ""
status: ""
risk: ""
description: ""
scenarios:
  - ""
examples: []
`
		filePath := filepath.Join(dirName, fileName)
		if err := os.WriteFile(filePath, []byte(emptyPlan), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		fmt.Printf("Successfully initialized '%s'\n", filePath)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVarP(&planName, "name", "n", "test-plan.yaml",
		"Name of the YAML file to create (.yaml is appended if missing)")
}
