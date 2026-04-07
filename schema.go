package main

import "fmt"

// TestPlan represents the YAML schema for a single test plan document.
type TestPlan struct {
	Feature     string     `yaml:"feature"`
	Type        string     `yaml:"type"`
	Status      string     `yaml:"status"`
	Issues      *string    `yaml:"issues,omitempty"`
	Docs        *string    `yaml:"docs,omitempty"`
	Description *string    `yaml:"description,omitempty"`
	Background  *string    `yaml:"background,omitempty"`
	Scenarios   []string   `yaml:"scenarios"`
	Examples    [][]string `yaml:"examples,omitempty"`
}

// ValidateSchema ensures the test plan uses accepted types and statuses.
func ValidateSchema(plan TestPlan) error {
	validTypes := map[string]bool{
		"functional":  true,
		"solution":    true,
		"performance": true,
		"reliability": true,
		"security":    true,
	}
	if !validTypes[plan.Type] {
		return fmt.Errorf("invalid type '%s': must be one of 'functional', 'solution', 'performance', 'reliability', or 'security'", plan.Type)
	}

	validStatuses := map[string]bool{
		"planned":     true,
		"implemented": true,
		"deprecated":  true,
	}
	if !validStatuses[plan.Status] {
		return fmt.Errorf("invalid status '%s': must be one of 'planned', 'implemented', or 'deprecated'", plan.Status)
	}
	return nil
}
