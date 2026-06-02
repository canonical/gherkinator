package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const multiDocYAML = `feature: "Job Submission"
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

func TestLoadTestPlans_Success(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(multiDocYAML), 0644)
	require.NoError(t, err)

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
	inputFile := filepath.Join(tmpDir, "bad.yaml")
	err := os.WriteFile(inputFile, []byte("{{{invalid yaml"), 0644)
	require.NoError(t, err)

	_, err = LoadTestPlans(inputFile)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode YAML")
}

func TestLoadTestPlans_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "empty.yaml")
	err := os.WriteFile(inputFile, []byte(""), 0644)
	require.NoError(t, err)

	plans, err := LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Empty(t, plans)
}

func TestWriteTestPlans_Success(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "out.yaml")

	plans := []TestPlan{
		{Feature: "Feature A", Type: "functional", Status: "planned", Risk: "stable", Scenarios: []string{"Scenario A"}},
		{Feature: "Feature B", Type: "security", Status: "implemented", Risk: "stable", Scenarios: []string{"Scenario B"}},
	}

	err := WriteTestPlans(outputFile, plans)
	require.NoError(t, err)

	// Read back and verify
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

func TestDeleteTestPlans_DeleteOne(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "GPU Job Submission", Type: "performance", Status: "implemented", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"Job Submission"})
	assert.Len(t, remaining, 2)
	assert.Len(t, deleted, 1)
	assert.Equal(t, "Job Submission", deleted[0])
	assert.Equal(t, "GPU Job Submission", remaining[0].Feature)
	assert.Equal(t, "Network Config", remaining[1].Feature)
}

func TestDeleteTestPlans_DeleteMultiple(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "GPU Job Submission", Type: "performance", Status: "implemented", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"Job Submission", "GPU Job Submission"})
	assert.Len(t, remaining, 1)
	assert.Len(t, deleted, 2)
	assert.Equal(t, "Network Config", remaining[0].Feature)
}

func TestDeleteTestPlans_CaseInsensitive(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"job submission"})
	assert.Len(t, remaining, 1)
	assert.Len(t, deleted, 1)
	assert.Equal(t, "Job Submission", deleted[0])
}

func TestDeleteTestPlans_NoMatch(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"nonexistent"})
	assert.Len(t, remaining, 1)
	assert.Empty(t, deleted)
}

func TestDeleteTestPlans_DeleteAll(t *testing.T) {
	plans := []TestPlan{
		{Feature: "Job Submission", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Network Config", Type: "security", Status: "planned", Risk: "stable"},
	}

	remaining, deleted := DeleteTestPlans(plans, []string{"Job Submission", "Network Config"})
	assert.Empty(t, remaining)
	assert.Len(t, deleted, 2)
}

func TestConfirmDeletion_UserConfirmsY(t *testing.T) {
	reader := strings.NewReader("Y\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.True(t, result)
}

func TestConfirmDeletion_UserDeclinesN(t *testing.T) {
	reader := strings.NewReader("n\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_UserDeclinesEmpty(t *testing.T) {
	reader := strings.NewReader("\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_UserDeclinesLowercaseY(t *testing.T) {
	reader := strings.NewReader("y\n")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_EmptyInput(t *testing.T) {
	reader := strings.NewReader("")
	result := ConfirmDeletion([]string{"Job Submission"}, reader)
	assert.False(t, result)
}

func TestConfirmDeletion_MultipleNames(t *testing.T) {
	reader := strings.NewReader("Y\n")
	result := ConfirmDeletion([]string{"Job Submission", "GPU Job Submission"}, reader)
	assert.True(t, result)
}

func TestQuoteNames(t *testing.T) {
	result := quoteNames([]string{"hello", "world"})
	assert.Equal(t, []string{`"hello"`, `"world"`}, result)
}

func TestDeleteCommand_WithYesFlag(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(multiDocYAML), 0644)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"delete", "-y", "-i", inputFile, "Job Submission"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// Verify the plan was deleted
	plans, err := LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 2)
	for _, p := range plans {
		assert.NotEqual(t, "Job Submission", p.Feature)
	}
}

func TestDeleteCommand_MultipleDeletes(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(multiDocYAML), 0644)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"delete", "-y", "-i", inputFile, "Job Submission", "GPU Job Submission"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	plans, err := LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.Equal(t, "Network Config", plans[0].Feature)
}

func TestDeleteCommand_NoMatch(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(multiDocYAML), 0644)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"delete", "-y", "-i", inputFile, "nonexistent"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// All plans should still be present
	plans, err := LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 3)
}

func TestDeleteCommand_MissingArgs(t *testing.T) {
	setupCommands()
	rootCmd.SetArgs([]string{"delete"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCommand_FileNotFound(t *testing.T) {
	setupCommands()
	rootCmd.SetArgs([]string{"delete", "-y", "-i", "/nonexistent/file.yaml", "something"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}
