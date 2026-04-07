package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateGherkin_ValidSyntax(t *testing.T) {
	gherkinText := `Feature: Login
  Scenario: User logs in
    Given a user exists
    When the user logs in
    Then the user sees the dashboard
`
	err := ValidateGherkin(gherkinText)
	assert.NoError(t, err)
}

func TestValidateGherkin_InvalidSyntax(t *testing.T) {
	gherkinText := `This is not valid gherkin at all`
	err := ValidateGherkin(gherkinText)
	assert.Error(t, err)
}

func TestProcessFile_GherkinFormat(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Login Feature"
type: "functional"
status: "planned"
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
	err = ProcessFile(inputFile, "gh", outputDir)
	assert.NoError(t, err)

	outFile := filepath.Join(outputDir, "login_feature.feature")
	assert.FileExists(t, outFile)

	content, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Feature: Login Feature")
	assert.Contains(t, string(content), "Scenario: User logs in")
}

func TestProcessFile_MarkdownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Login Feature"
type: "functional"
status: "planned"
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
	err = ProcessFile(inputFile, "md", outputDir)
	assert.NoError(t, err)

	outFile := filepath.Join(outputDir, "login_feature.md")
	assert.FileExists(t, outFile)

	content, err := os.ReadFile(outFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Login Feature")
}

func TestProcessFile_MultiDocument(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Feature One"
type: "functional"
status: "planned"
scenarios:
  - |
    Scenario One
    Given something
    When action
    Then result
---
feature: "Feature Two"
type: "security"
status: "implemented"
scenarios:
  - |
    Scenario Two
    Given another thing
    When another action
    Then another result
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "md", outputDir)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "feature_one.md"))
	assert.FileExists(t, filepath.Join(outputDir, "feature_two.md"))
}

func TestProcessFile_InvalidSchema(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Bad Feature"
type: "invalid_type"
status: "planned"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "gh", outputDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

func TestProcessFile_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `{{{not valid yaml`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "gh", outputDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode YAML")
}

func TestProcessFile_FileNotFound(t *testing.T) {
	err := ProcessFile("/nonexistent/file.yaml", "gh", "/tmp/out")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestProcessFile_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test"
type: "functional"
status: "planned"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "xml", outputDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestProcessFile_EmptyFeatureName(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: ""
type: "functional"
status: "planned"
scenarios:
  - |
    A scenario
    Given x
    When y
    Then z
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "md", outputDir)
	assert.NoError(t, err)

	// Should use plan_1 as fallback filename
	assert.FileExists(t, filepath.Join(outputDir, "plan_1.md"))
}
