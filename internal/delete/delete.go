// Package delete implements the `gherkinator delete` subcommand: it removes
// selected test plan entries from a multi-document YAML file.
package delete

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"gherkinator/internal/common"
)

// DeleteTestPlans removes test plans whose feature names match the provided
// list (case-insensitive). It returns the remaining plans and the names that
// were actually found and deleted.
func DeleteTestPlans(plans []common.TestPlan, featureNames []string) (remaining []common.TestPlan, deleted []string) {
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
