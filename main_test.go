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
	rootCmd.SetArgs([]string{"generate", "--format", "gh", "-o", outputDir, inputFile})
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
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, inputFile})
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
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, "--risk", "edge", inputFile})
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
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, "--risk", "stable", inputFile})
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
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, "--risk", "invalid", inputFile})
	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--risk must be one of")
}

func TestGenerateCommand_FileNotFound(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	rootCmd.SetArgs([]string{"generate", "--format", "gh", "-o", tmpDir, filepath.Join(tmpDir, "nonexistent.yaml")})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestGenerateCommand_MultipleFiles(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()

	file1 := filepath.Join(tmpDir, "alpha.yaml")
	require.NoError(t, os.WriteFile(file1, []byte(`feature: "Alpha Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    Alpha scenario
    Given something
    When action
    Then result
`), 0644))

	file2 := filepath.Join(tmpDir, "beta.yaml")
	require.NoError(t, os.WriteFile(file2, []byte(`feature: "Beta Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Beta scenario
    Given something
    When action
    Then result
`), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, file1, file2})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "alpha_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "beta_feature.md"))
}

func TestGenerateCommand_DirectoryArg(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	inputDir := filepath.Join(tmpDir, "plans")
	require.NoError(t, os.MkdirAll(inputDir, 0755))

	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "gamma.yaml"), []byte(`feature: "Gamma Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    Gamma scenario
    Given something
    When action
    Then result
`), 0644))

	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "delta.yml"), []byte(`feature: "Delta Feature"
type: "performance"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Delta scenario
    Given something
    When action
    Then result
`), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, inputDir})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "gamma_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "delta_feature.md"))
}

func TestGenerateCommand_MixedFilesAndDirectories(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()

	inputDir := filepath.Join(tmpDir, "plans")
	require.NoError(t, os.MkdirAll(inputDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(inputDir, "epsilon.yaml"), []byte(`feature: "Epsilon Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    Epsilon scenario
    Given x
`), 0644))

	explicitFile := filepath.Join(tmpDir, "zeta.yaml")
	require.NoError(t, os.WriteFile(explicitFile, []byte(`feature: "Zeta Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Zeta scenario
    Given x
`), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, inputDir, explicitFile})
	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "epsilon_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "zeta_feature.md"))
}

func TestGenerateCommand_DefaultScansCwd(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "theta.yaml"), []byte(`feature: "Theta Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    Theta scenario
    Given x
`), 0644))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "theta_feature.md"))
}

func TestGenerateCommand_NoFlagsScansCwd(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "iota.yaml"), []byte(`feature: "Iota Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    Iota scenario
    Given x
`), 0644))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	rootCmd.SetArgs([]string{"generate"})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(tmpDir, "iota_feature.feature"))
}

func TestServeCommand_NoYAMLFiles(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	rootCmd.SetArgs([]string{"serve"})
	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no YAML files found")
}

func TestServeCommand_FileNotFound(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()
	rootCmd.SetArgs([]string{"serve", filepath.Join(tmpDir, "nonexistent.yaml")})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}
