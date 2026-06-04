package edit

import (
	"gherkinator/internal/common"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextEditor_UsesVISUAL(t *testing.T) {
	t.Setenv("VISUAL", "cat")
	t.Setenv("EDITOR", "")

	content := []byte("test content")
	edited, err := TextEditor(content)
	require.NoError(t, err)
	assert.Equal(t, content, edited)
}

func TestTextEditor_UsesEDITOR(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "cat")

	content := []byte("test content")
	edited, err := TextEditor(content)
	require.NoError(t, err)
	assert.Equal(t, content, edited)
}

func TestTextEditor_NoEditorAvailable(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	path := os.Getenv("PATH")
	if err := os.Setenv("PATH", "/nonexistent"); err != nil {
		t.Fatalf("failed to set PATH: %v", err)
	}

	content := []byte("test content")
	_, err := TextEditor(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no text editor found")

	if err := os.Setenv("PATH", path); err != nil {
		t.Fatalf("failed to restore PATH: %v", err)
	}
}

func TestEditHelpTemplate_ContainsSchemaFields(t *testing.T) {
	template := EditHelpTemplate()
	assert.Contains(t, template, "feature:")
	assert.Contains(t, template, "type:")
	assert.Contains(t, template, "status:")
	assert.Contains(t, template, "risk:")
	assert.Contains(t, template, "scenarios:")
	assert.Contains(t, template, "examples:")
}

func TestEditHelpTemplate_ContainsValidTypes(t *testing.T) {
	template := EditHelpTemplate()
	assert.Contains(t, template, "functional")
	assert.Contains(t, template, "solution")
	assert.Contains(t, template, "performance")
	assert.Contains(t, template, "reliability")
	assert.Contains(t, template, "security")
}

func TestEditHelpTemplate_ContainsValidStatuses(t *testing.T) {
	template := EditHelpTemplate()
	assert.Contains(t, template, "planned")
	assert.Contains(t, template, "implemented")
	assert.Contains(t, template, "deprecated")
}

func TestEditHelpTemplate_ContainsValidRisks(t *testing.T) {
	template := EditHelpTemplate()
	assert.Contains(t, template, "edge")
	assert.Contains(t, template, "beta")
	assert.Contains(t, template, "candidate")
	assert.Contains(t, template, "stable")
}

func TestValidateEditContent_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: "Test Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Test scenario"
`
	err := ValidateEditContent(filename, []byte(yamlContent))
	require.NoError(t, err)

	loaded, err := common.LoadTestPlans(filename)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, "Test Feature", loaded[0].Feature)
}

func TestValidateEditContent_InvalidType(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: "Test Feature"
type: "invalid_type"
status: "planned"
risk: "stable"
scenarios:
  - "Test scenario"
`
	err := ValidateEditContent(filename, []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestValidateEditContent_InvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: "Test Feature"
type: "functional"
status: "invalid_status"
risk: "stable"
scenarios:
  - "Test scenario"
`
	err := ValidateEditContent(filename, []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestValidateEditContent_EmptyFeature(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: ""
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Test scenario"
`
	err := ValidateEditContent(filename, []byte(yamlContent))
	require.NoError(t, err)

	loaded, err := common.LoadTestPlans(filename)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
}

func TestValidateEditContent_MultiplePlans(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: "Feature 1"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Scenario 1"
---
feature: "Feature 2"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Scenario 2"
`
	err := ValidateEditContent(filename, []byte(yamlContent))
	require.NoError(t, err)

	loaded, err := common.LoadTestPlans(filename)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "Feature 1", loaded[0].Feature)
	assert.Equal(t, "Feature 2", loaded[1].Feature)
}

func TestValidateEditContent_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := "{{{invalid yaml"
	err := ValidateEditContent(filename, []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YAML parse error")
}

func TestValidateEditContent_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	err := ValidateEditContent(filename, []byte{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid test plans found")
}

func TestValidateEditContent_FileError(t *testing.T) {
	yamlContent := `feature: "Test"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Scenario"
`
	err := ValidateEditContent("/nonexistent/dir/file.yaml", []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")
}
