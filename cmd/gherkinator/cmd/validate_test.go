package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommand_ValidSingleFile(t *testing.T) {
	resetFlags()
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
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	rootCmd.SetArgs([]string{"validate", inputFile})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestValidateCommand_ValidMultipleDocs(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Feature A"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Scenario A"
---
feature: "Feature B"
type: "security"
status: "implemented"
risk: "candidate"
scenarios:
  - "Scenario B"
---
feature: "Feature C"
type: "performance"
status: "deprecated"
risk: "edge"
scenarios:
  - "Scenario C"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	rootCmd.SetArgs([]string{"validate", inputFile})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestValidateCommand_InvalidType(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test"
type: "bogus_type"
status: "planned"
risk: "stable"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	rootCmd.SetArgs([]string{"validate", inputFile})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test plan '"+inputFile+"' failed validation (document 1): invalid type")
}

func TestValidateCommand_InvalidStatus(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test"
type: "functional"
status: "bogus_status"
risk: "stable"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	rootCmd.SetArgs([]string{"validate", inputFile})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test plan '"+inputFile+"' failed validation (document 1): invalid status")
}

func TestValidateCommand_InvalidRisk(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test"
type: "functional"
status: "planned"
risk: "bogus_risk"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	rootCmd.SetArgs([]string{"validate", inputFile})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test plan '"+inputFile+"' failed validation (document 1): invalid risk")
}

func TestValidateCommand_BadYAMLSyntax(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "bad.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte("{{{invalid yaml"), 0644))

	rootCmd.SetArgs([]string{"validate", inputFile})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read '"+inputFile+"'")
}

func TestValidateCommand_FileNotFound(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	missing := filepath.Join(tmpDir, "nonexistent.yaml")

	rootCmd.SetArgs([]string{"validate", missing})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestValidateCommand_MultipleFiles(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "alpha.yaml")
	require.NoError(t, os.WriteFile(file1, []byte(`feature: "Alpha Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Alpha scenario"
`), 0644))

	file2 := filepath.Join(tmpDir, "beta.yaml")
	require.NoError(t, os.WriteFile(file2, []byte(`feature: "Beta Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Beta scenario"
`), 0644))

	rootCmd.SetArgs([]string{"validate", file1, file2})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestValidateCommand_MixedValidInvalid(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()

	validFile := filepath.Join(tmpDir, "valid.yaml")
	require.NoError(t, os.WriteFile(validFile, []byte(`feature: "Valid Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "valid scenario"
`), 0644))

	invalidFile := filepath.Join(tmpDir, "invalid.yaml")
	require.NoError(t, os.WriteFile(invalidFile, []byte(`feature: "Invalid Feature"
type: "functional"
status: "planned"
risk: "bogus_risk"
scenarios:
  - "invalid scenario"
`), 0644))

	rootCmd.SetArgs([]string{"validate", validFile, invalidFile})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "test plan '"+invalidFile+"' failed validation (document 1)")
	assert.Contains(t, err.Error(), "invalid risk")
}

func TestValidateCommand_DirectoryArg(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "plans")
	require.NoError(t, os.MkdirAll(inputDir, 0755))

	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "gamma.yaml"), []byte(`feature: "Gamma Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Gamma scenario"
`), 0644))

	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "delta.yml"), []byte(`feature: "Delta Feature"
type: "performance"
status: "implemented"
risk: "stable"
scenarios:
  - "Delta scenario"
`), 0644))

	rootCmd.SetArgs([]string{"validate", inputDir})
	err := rootCmd.Execute()
	assert.NoError(t, err)
}

func TestValidateCommand_NoArgsScansCwd(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "kappa.yaml"), []byte(`feature: "Kappa Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Kappa scenario"
`), 0644))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	rootCmd.SetArgs([]string{"validate"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
}

func TestValidateCommand_NoYAMLFound(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte(""), 0644))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	rootCmd.SetArgs([]string{"validate"})
	err = rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no YAML files found")
}

func TestValidateCommand_MultiDocSecondDocInvalid(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Valid Plan"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "valid scenario"
---
feature: "Invalid Plan"
type: "functional"
status: "planned"
risk: "stable"
scenarios: []
---
feature: "Another Invalid Plan"
type: "security"
status: "planned"
risk: "bogus"
scenarios:
  - "another scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	rootCmd.SetArgs([]string{"validate", inputFile})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "document 3")
}
