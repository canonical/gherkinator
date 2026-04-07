package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"go.yaml.in/yaml/v3"
)

// LoadTestPlans reads a multi-document YAML file and returns all TestPlan entries.
func LoadTestPlans(filename string) ([]TestPlan, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var plans []TestPlan
	decoder := yaml.NewDecoder(file)
	for {
		var plan TestPlan
		if err := decoder.Decode(&plan); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}
		plans = append(plans, plan)
	}
	return plans, nil
}

// WriteTestPlans writes a slice of TestPlan entries to a YAML file as a
// multi-document stream separated by "---".
func WriteTestPlans(filename string, plans []TestPlan) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	encoder := yaml.NewEncoder(file)
	defer func() {
		_ = encoder.Close()
	}()

	for _, plan := range plans {
		if err := encoder.Encode(&plan); err != nil {
			return fmt.Errorf("failed to encode YAML: %w", err)
		}
	}
	return nil
}

// DeleteTestPlans removes test plans whose feature names match the provided
// list (case-insensitive). It returns the remaining plans and the names that
// were actually found and deleted.
func DeleteTestPlans(plans []TestPlan, featureNames []string) (remaining []TestPlan, deleted []string) {
	toDelete := make(map[string]bool)
	for _, name := range featureNames {
		toDelete[strings.ToLower(name)] = true
	}

	for _, plan := range plans {
		if toDelete[strings.ToLower(plan.Feature)] {
			deleted = append(deleted, plan.Feature)
		} else {
			remaining = append(remaining, plan)
		}
	}
	return remaining, deleted
}

// ConfirmDeletion prompts the user for confirmation and reads from the
// provided reader. Returns true only when the user enters "Y".
func ConfirmDeletion(featureNames []string, reader io.Reader) bool {
	//nolint:errcheck // Writing to stdout; error is not actionable
	fmt.Fprintf(os.Stdout, "Are you sure you want to delete test plans %s? [Y/n] ",
		strings.Join(quoteNames(featureNames), ", "))

	scanner := bufio.NewScanner(reader)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()) == "Y"
	}
	return false
}

// quoteNames wraps each name in double quotes for display.
func quoteNames(names []string) []string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = fmt.Sprintf("%q", name)
	}
	return quoted
}
