package common

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTestPlans_Success(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/test-plan.yaml"
	yamlContent := `feature: "Job Submission"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Submit a job"
---
feature: "GPU Job Submission"
type: "performance"
status: "implemented"
risk: "stable"
scenarios:
  - "Submit a GPU job"
---
feature: "Network Config"
type: "security"
status: "planned"
risk: "stable"
scenarios:
  - "Configure network"
`
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	plans, err := LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 3)
	assert.Equal(t, "Job Submission", plans[0].Feature)
	assert.Equal(t, "GPU Job Submission", plans[1].Feature)
	assert.Equal(t, "Network Config", plans[2].Feature)
}

func TestLoadTestPlans_FileNotFound(t *testing.T) {
	_, err := LoadTestPlans("/nonexistent/path.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestLoadTestPlans_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/bad.yaml"
	require.NoError(t, os.WriteFile(inputFile, []byte("{{{invalid yaml"), 0644))

	_, err := LoadTestPlans(inputFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode YAML")
}

func TestLoadTestPlans_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := tmpDir + "/empty.yaml"
	require.NoError(t, os.WriteFile(inputFile, []byte(""), 0644))

	plans, err := LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Empty(t, plans)
}

func TestWriteTestPlans_Success(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := tmpDir + "/out.yaml"

	plans := []TestPlan{
		{Feature: "Feature A", Type: "functional", Status: "planned", Risk: "stable", Scenarios: []string{"Scenario A"}},
		{Feature: "Feature B", Type: "security", Status: "implemented", Risk: "stable", Scenarios: []string{"Scenario B"}},
	}

	err := WriteTestPlans(outputFile, plans)
	require.NoError(t, err)

	loaded, err := LoadTestPlans(outputFile)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "Feature A", loaded[0].Feature)
	assert.Equal(t, "Feature B", loaded[1].Feature)
}

func TestWriteTestPlans_InvalidPath(t *testing.T) {
	err := WriteTestPlans("/nonexistent/dir/file.yaml", []TestPlan{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")
}
