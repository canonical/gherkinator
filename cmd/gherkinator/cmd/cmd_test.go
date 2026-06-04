package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gherkinator/internal/common"
)

// resetFlags restores the package-level cobra flag variables to their
// default values. This mirrors the idempotent setupCommands() helper from
// the original main.go and ensures tests do not leak flag state to one
// another.
func resetFlags() {
	outputDir = "."
	format = "gh"
	riskFilter = ""
	statusFilter = ""
	serveName = ""
	skipConfirm = false
	deleteInputFile = ""
	cleanDir = "."

	// Reset args and flags changed for the rootCmd itself.
	rootCmd.SetArgs([]string{})
}

func TestRootCmd_HasSubcommands(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "gherkinator", rootCmd.Use)
	assert.Equal(t, "A testing plan management and generation tool", rootCmd.Short)
	assert.Len(t, rootCmd.Commands(), 7)
}

func TestInitCommand_CreatesDirectoryAndFile(t *testing.T) {
	resetFlags()
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
	resetFlags()
	rootCmd.SetArgs([]string{"init"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestInitCommand_CustomName(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "my-test-plans")

	rootCmd.SetArgs([]string{"init", targetDir, "-n", "my-plan.yaml"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	filePath := filepath.Join(targetDir, "my-plan.yaml")
	assert.FileExists(t, filePath)
}

func TestInitCommand_NameWithoutExtension(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "my-test-plans")

	rootCmd.SetArgs([]string{"init", targetDir, "-n", "my-plan"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	filePath := filepath.Join(targetDir, "my-plan.yaml")
	assert.FileExists(t, filePath)
}

func TestInitCommand_NameWithYmlExtension(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "my-test-plans")

	rootCmd.SetArgs([]string{"init", targetDir, "-n", "my-plan.yml"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	filePath := filepath.Join(targetDir, "my-plan.yml")
	assert.FileExists(t, filePath)
	assert.NoFileExists(t, filepath.Join(targetDir, "my-plan.yml.yaml"))
}

func TestInitCommand_NameWithYAMLExtensionCaseInsensitive(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "my-test-plans")

	rootCmd.SetArgs([]string{"init", targetDir, "-n", "Plan.YAML"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	filePath := filepath.Join(targetDir, "Plan.YAML")
	assert.FileExists(t, filePath)
	assert.NoFileExists(t, filepath.Join(targetDir, "Plan.YAML.yaml"))
}

func TestInitCommand_EmptyName(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	targetDir := filepath.Join(tmpDir, "my-test-plans")

	rootCmd.SetArgs([]string{"init", targetDir, "-n", ""})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--name cannot be empty")
}

func TestGenerateCommand_InvalidFormat(t *testing.T) {
	resetFlags()
	rootCmd.SetArgs([]string{"generate", "--format", "xml"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestGenerateCommand_GherkinFormat(t *testing.T) {
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
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "gh", "-o", outputDir, inputFile})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "test_feature.feature"))
}

func TestGenerateCommand_MarkdownFormat(t *testing.T) {
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
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, inputFile})
	err = rootCmd.Execute()
	assert.NoError(t, err)
	assert.FileExists(t, filepath.Join(outputDir, "test_feature.md"))
}

func TestGenerateCommand_RiskFilter(t *testing.T) {
	resetFlags()
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
	resetFlags()
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
	resetFlags()
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
	resetFlags()
	tmpDir := t.TempDir()
	rootCmd.SetArgs([]string{"generate", "--format", "gh", "-o", tmpDir, filepath.Join(tmpDir, "nonexistent.yaml")})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestGenerateCommand_MultipleFiles(t *testing.T) {
	resetFlags()
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
	resetFlags()
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
	resetFlags()
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
	resetFlags()
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
	resetFlags()
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
	resetFlags()
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
	resetFlags()
	tmpDir := t.TempDir()
	rootCmd.SetArgs([]string{"serve", filepath.Join(tmpDir, "nonexistent.yaml")})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestDeleteCommand_WithYesFlag(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Job Submission"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Submit a job"
---
feature: "Network Config"
type: "security"
status: "planned"
risk: "stable"
scenarios:
  - "Configure network"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"delete", "-y", "-i", inputFile, "Job Submission"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// Verify the plan was deleted
	plans, err := common.LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 1)
	for _, p := range plans {
		assert.NotEqual(t, "Job Submission", p.Feature)
	}
}

func TestDeleteCommand_MultipleDeletes(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
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
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"delete", "-y", "-i", inputFile, "Job Submission", "GPU Job Submission"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	plans, err := common.LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.Equal(t, "Network Config", plans[0].Feature)
}

func TestDeleteCommand_NoMatch(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Job Submission"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Submit a job"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	rootCmd.SetArgs([]string{"delete", "-y", "-i", inputFile, "nonexistent"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// All plans should still be present
	plans, err := common.LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 1)
}

func TestDeleteCommand_MissingArgs(t *testing.T) {
	resetFlags()
	rootCmd.SetArgs([]string{"delete"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestDeleteCommand_MissingInputFlag(t *testing.T) {
	resetFlags()
	rootCmd.SetArgs([]string{"delete", "Job Submission"})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--input/-i is required")
}

func TestDeleteCommand_FileNotFound(t *testing.T) {
	resetFlags()
	rootCmd.SetArgs([]string{"delete", "-y", "-i", "/nonexistent/file.yaml", "something"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestCleanCommand_Success(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()

	// Create files to clean
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("feature"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs.md"), []byte("docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test-plan.yaml"), []byte("yaml"), 0644))

	rootCmd.SetArgs([]string{"clean", "-d", tmpDir})
	err := rootCmd.Execute()
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "docs.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "test-plan.yaml"))
}

func TestCleanCommand_NonexistentDir(t *testing.T) {
	resetFlags()
	rootCmd.SetArgs([]string{"clean", "-d", "/nonexistent/dir"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestEditCommand_UpdatesFile(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	original := `feature: "Old Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Old scenario"
`
	require.NoError(t, os.WriteFile(inputFile, []byte(original), 0644))

	// `cat` echoes the buffer unchanged, then ValidateEditContent
	// re-serialises the YAML back to disk.
	t.Setenv("VISUAL", "cat")
	rootCmd.SetArgs([]string{"edit", inputFile})
	err := rootCmd.Execute()
	require.NoError(t, err)

	plans, err := common.LoadTestPlans(inputFile)
	require.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.Equal(t, "Old Feature", plans[0].Feature)
}

func TestEditCommand_RejectsInvalidYAML(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte("feature: \"Original\"\ntype: \"functional\"\nstatus: \"planned\"\nrisk: \"stable\"\nscenarios:\n  - \"Original scenario\"\n"), 0644))

	// A custom editor that exits 0 but leaves the buffer empty, so
	// ValidateEditContent runs and reports "no valid test plans found".
	emptyEditor := `#!/bin/sh
: > "$1"
exit 0
`
	editorPath := filepath.Join(tmpDir, "empty_editor.sh")
	require.NoError(t, os.WriteFile(editorPath, []byte(emptyEditor), 0755))
	t.Setenv("VISUAL", editorPath)

	rootCmd.SetArgs([]string{"edit", inputFile})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid test plans found")
}

func TestEditCommand_MissingArgs(t *testing.T) {
	resetFlags()
	rootCmd.SetArgs([]string{"edit"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestEditCommand_FileNotFound(t *testing.T) {
	resetFlags()
	t.Setenv("VISUAL", "cat")
	rootCmd.SetArgs([]string{"edit", "/nonexistent/file.yaml"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestGenerateCommand_StatusFilter(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Planned Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Planned scenario"
---
feature: "Implemented Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Implemented scenario"
---
feature: "Deprecated Feature"
type: "solution"
status: "deprecated"
risk: "stable"
scenarios:
  - "Deprecated scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, "--status", "planned", inputFile})
	err := rootCmd.Execute()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "planned_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "implemented_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "deprecated_feature.md"))
}

func TestGenerateCommand_StatusFilter_Implemented(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Planned Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Planned scenario"
---
feature: "Implemented Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Implemented scenario"
---
feature: "Deprecated Feature"
type: "solution"
status: "deprecated"
risk: "stable"
scenarios:
  - "Deprecated scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, "--status", "implemented", inputFile})
	err := rootCmd.Execute()
	require.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(outputDir, "planned_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "implemented_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "deprecated_feature.md"))
}

func TestGenerateCommand_InvalidStatusFlag(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir, "--status", "invalid", inputFile})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--status must be one of")
}

func TestGenerateCommand_StatusAndRisk_Intersection(t *testing.T) {
	// --status=implemented --risk=candidate: only "implemented" plans whose
	// risk is edge, beta, or candidate.
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Implemented Edge"
type: "functional"
status: "implemented"
risk: "edge"
scenarios:
  - "Edge scenario"
---
feature: "Implemented Candidate"
type: "security"
status: "implemented"
risk: "candidate"
scenarios:
  - "Candidate scenario"
---
feature: "Implemented Stable"
type: "solution"
status: "implemented"
risk: "stable"
scenarios:
  - "Stable scenario"
---
feature: "Planned Beta"
type: "functional"
status: "planned"
risk: "beta"
scenarios:
  - "Planned scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir,
		"--status", "implemented", "--risk", "candidate", inputFile})
	err := rootCmd.Execute()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "implemented_edge.md"))
	assert.FileExists(t, filepath.Join(outputDir, "implemented_candidate.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "implemented_stable.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "planned_beta.md"))
}

func TestGenerateCommand_StatusAndStable_Intersection(t *testing.T) {
	// --risk=stable matches every risk level, so the intersection with
	// --status=implemented should yield only implemented plans.
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Implemented Edge"
type: "functional"
status: "implemented"
risk: "edge"
scenarios:
  - "Edge scenario"
---
feature: "Implemented Stable"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Stable scenario"
---
feature: "Planned Stable"
type: "solution"
status: "planned"
risk: "stable"
scenarios:
  - "Planned scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	rootCmd.SetArgs([]string{"generate", "--format", "md", "-o", outputDir,
		"--status", "implemented", "--risk", "stable", inputFile})
	err := rootCmd.Execute()
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "implemented_edge.md"))
	assert.FileExists(t, filepath.Join(outputDir, "implemented_stable.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "planned_stable.md"))
}

func TestServeCommand_InvalidStatusFlag(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	yamlContent := `feature: "Test"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	// --status validation happens before any disk I/O, so the YAML file
	// can be missing or invalid; we just need serve to refuse the flag.
	rootCmd.SetArgs([]string{"serve", "--status", "bogus", inputFile})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--status must be one of")
}
