package common

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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
###     When a job is submitted
###     Then the job completes successfully
### examples:
###   - - alice
###     - admin
###   - - bob
###     - viewer
###
### Multiple plans are separated by '---'.
###
`
}

// TextEditor opens a text editor with a temporary YAML file for editing.
func TextEditor(content []byte) ([]byte, error) {
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
		if editor == "" {
			for _, p := range []string{"editor", "vi", "emacs", "nano"} {
				_, err := exec.LookPath(p)
				if err == nil {
					editor = p
					break
				}
			}
			if editor == "" {
				return nil, errors.New("no text editor found, please set the EDITOR environment variable")
			}
		}
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
	var plans []TestPlan
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	for {
		var plan TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("YAML parse error: %w", err)
		}
		if err := ValidateSchema(plan); err != nil {
			return err
		}
		plans = append(plans, plan)
	}

	if len(plans) == 0 {
		return fmt.Errorf("no valid test plans found in input")
	}

	return WriteTestPlans(filename, plans)
}
