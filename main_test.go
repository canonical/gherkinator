package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_HasSubcommands(t *testing.T) {
	setupCommands()
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "gherkinator", rootCmd.Use)
	assert.Equal(t, "A testing plan management and generation tool", rootCmd.Short)
	assert.Len(t, rootCmd.Commands(), 5)
}

func TestInitCommand_CreatesDirectoryAndFile(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "my-test-plans")

	rootCmd.SetArgs([]string{"init", targetDir})
	err := rootCmd.Execute()
	require.NoError(t, err)

	filePath := filepath.Join(targetDir, "test-plan.yaml")
	assert.FileExists(t, filePath)

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "feature:")
	assert.Contains(t, string(content), "type:")
	assert.Contains(t, string(content), "status:")
	assert.Contains(t, string(content), "risk:")
	assert.Contains(t, string(content), "scenarios:")
}

func TestInitCommand_MissingArgs(t *testing.T) {
	setupCommands()
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestGenerateCommand_InvalidFormat(t *testing.T) {
	setupCommands()
	rootCmd.SetArgs([]string{"generate", "--format", "xml"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestGenerateCommand_GherkinFormat(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
    When the user logs in
    Then the user sees the dashboard
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "gh", "-i", inputFile, "-o", outputDir})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "test_feature.feature"))
}

func TestGenerateCommand_MarkdownFormat(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
    When the user logs in
    Then the user sees the dashboard
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-i", inputFile, "-o", outputDir})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "test_feature.md"))
}

func TestGenerateCommand_RiskFilter(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Edge Feature"
type: "functional"
status: "planned"
risk: "edge"
scenarios:
  - |
    Edge scenario
    Given something
---
feature: "Stable Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Stable scenario
    Given something else
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-i", inputFile, "-o", outputDir, "--risk", "edge"})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "edge_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "stable_feature.md"))
}

func TestGenerateCommand_RiskFilterStable(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Edge Feature"
type: "functional"
status: "planned"
risk: "edge"
scenarios:
  - |
    Edge scenario
    Given something
---
feature: "Stable Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Stable scenario
    Given something else
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-i", inputFile, "-o", outputDir, "--risk", "stable"})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "edge_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "stable_feature.md"))
}

func TestGenerateCommand_InvalidRiskFlag(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-i", inputFile, "-o", outputDir, "--risk", "invalid"})
	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--risk must be one of")
}

func TestGenerateCommand_FileNotFound(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	rootCmd.SetArgs([]string{"generate", "--format", "gh", "-i", filepath.Join(tmpDir, "nonexistent.yaml"), "-o", tmpDir})
	err := rootCmd.Execute()
	assert.Error(t, err)
}
