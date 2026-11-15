package edit

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/canonical/gherkinator/internal/common"

	"go.yaml.in/yaml/v3"
)

// EditHelpTemplate returns a YAML comment header explaining the test plan schema.
func EditHelpTemplate() string {
	return `### This is a YAML representation of your test plan(s).
### Any line starting with '###' will be ignored.
###
### A sample test plan looks like:
### feature: "GPU job submission"
### type: functional            # One of: functional | solution | performance
###                             #         reliability | security
### status: planned             # One of: planned | implemented | deprecated
### risk: stable                # One of: edge | beta | candidate | stable
### description: "Test GPU jobs"
### background: |-
###   Given the cluster is available
###   And I am logged in as a user
### scenarios:
###   - |-
###     Submit a job
###     Given the system is running
###     When user <user> submits a job
###     Then the job completes successfully
### examples:
###   - - alice
###   - - bob
### tags:                       # Optional: deployment classification tags
###   - single-node            # (non-empty, no whitespace, no commas)
###
### Multiple plans are separated by '---'.
###
`
}

// defaultEditors is the ordered fallback list searched when $VISUAL and
// $EDITOR are unset or unusable. It finds the environment's default editor,
// which inside the strictly confined snap is the bundled nano.
var defaultEditors = []string{"editor", "vi", "emacs", "nano"}

// resolveEditor returns the editor command line to launch. $VISUAL wins over
// $EDITOR; a configured editor that cannot be resolved produces a warning on
// stderr and the next candidate is tried.
func resolveEditor() (string, error) {
	for _, env := range []string{"VISUAL", "EDITOR"} {
		editor := strings.TrimSpace(os.Getenv(env))
		if editor == "" {
			continue
		}
		parts := strings.Fields(editor)
		if _, err := exec.LookPath(parts[0]); err == nil {
			return editor, nil
		}
		fmt.Fprintf(os.Stderr, "warning: editor %q from $%s is not accessible in this environment; falling back to the default editor\n", editor, env) //nolint:errcheck // Writing to stderr; error is not actionable
	}

	for _, candidate := range defaultEditors {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("no text editor found, please set the EDITOR environment variable")
}

// TextEditor opens a text editor with a temporary YAML file for editing.
func TextEditor(content []byte) ([]byte, error) {
	editor, err := resolveEditor()
	if err != nil {
		return nil, err
	}

	f, err := os.CreateTemp("", "gherkinator_edit_*.yaml")
	if err != nil {
		return nil, err
	}
	path := f.Name()

	if _, err := f.Write(content); err != nil {
		_ = f.Close()
		return nil, err
	}
	if err := f.Close(); err != nil {
		return nil, err
	}

	cmdParts := strings.Fields(editor)
	cmd := exec.Command(cmdParts[0], append(cmdParts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("editor exited with error: %w", err)
	}

	content, err = os.ReadFile(path)
	_ = os.Remove(path)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// ValidateEditContent validates the YAML content as test plans and writes
// it to the file if valid.
func ValidateEditContent(filename string, content []byte) error {
	var plans []common.TestPlan
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	for {
		var plan common.TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("YAML parse error: %w", err)
		}
		if err := common.ValidateSchema(plan); err != nil {
			return err
		}
		plans = append(plans, plan)
	}

	if len(plans) == 0 {
		return fmt.Errorf("no valid test plans found in input")
	}

	return common.WriteTestPlans(filename, plans)
}
