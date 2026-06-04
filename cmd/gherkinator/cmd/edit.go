package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"gherkinator/internal/common"
)

// editCmd opens the user's text editor pre-populated with a YAML test plan
// file's contents (and a schema reference comment block).  On save, the
// edited buffer is parsed and validated; valid content is written back to
// the original file.
var editCmd = &cobra.Command{
	Use:   "edit [file]",
	Short: "Edit a test plan YAML file in $VISUAL/$EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]
		original, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", filename, err)
		}

		// Prepend the schema help template so the user can see the
		// valid values for each field while editing.
		buffer := append([]byte(common.EditHelpTemplate()), original...)

		edited, err := common.TextEditor(buffer)
		if err != nil {
			return err
		}

		if err := common.ValidateEditContent(filename, edited); err != nil {
			return err
		}
		fmt.Printf("Successfully updated '%s'\n", filename)
		return nil
	},
}
