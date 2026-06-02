package common

import (
	"fmt"
	"io"
	"os"

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
