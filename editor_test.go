package main

import (
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
	os.Setenv("PATH", "/nonexistent")

	content := []byte("test content")
	_, err := TextEditor(content)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no text editor found")

	os.Setenv("PATH", path)
}

func TestEditHelpTemplate_ContainsSchemaFields(t *testing.T) {
	template := editHelpTemplate()
	assert.Contains(t, template, "feature:")
	assert.Contains(t, template, "type:")
	assert.Contains(t, template, "status:")
	assert.Contains(t, template, "scenarios:")
	assert.Contains(t, template, "examples:")
}

func TestEditHelpTemplate_ContainsValidTypes(t *testing.T) {
	template := editHelpTemplate()
	assert.Contains(t, template, "functional")
	assert.Contains(t, template, "solution")
	assert.Contains(t, template, "performance")
	assert.Contains(t, template, "reliability")
	assert.Contains(t, template, "security")
}

func TestEditHelpTemplate_ContainsValidStatuses(t *testing.T) {
	template := editHelpTemplate()
	assert.Contains(t, template, "planned")
	assert.Contains(t, template, "implemented")
	assert.Contains(t, template, "deprecated")
}

func TestValidateEditContent_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: "Test Feature"
type: "functional"
status: "planned"
scenarios:
  - "Test scenario"
`
	err := validateEditContent(filename, []byte(yamlContent))
	require.NoError(t, err)

	loaded, err := LoadTestPlans(filename)
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
scenarios:
  - "Test scenario"
`
	err := validateEditContent(filename, []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestValidateEditContent_InvalidStatus(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: "Test Feature"
type: "functional"
status: "invalid_status"
scenarios:
  - "Test scenario"
`
	err := validateEditContent(filename, []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestValidateEditContent_EmptyFeature(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: ""
type: "functional"
status: "planned"
scenarios:
  - "Test scenario"
`
	err := validateEditContent(filename, []byte(yamlContent))
	require.NoError(t, err)

	loaded, err := LoadTestPlans(filename)
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
}

func TestValidateEditContent_MultiplePlans(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := `feature: "Feature 1"
type: "functional"
status: "planned"
scenarios:
  - "Scenario 1"
---
feature: "Feature 2"
type: "security"
status: "implemented"
scenarios:
  - "Scenario 2"
`
	err := validateEditContent(filename, []byte(yamlContent))
	require.NoError(t, err)

	loaded, err := LoadTestPlans(filename)
	require.NoError(t, err)
	assert.Len(t, loaded, 2)
	assert.Equal(t, "Feature 1", loaded[0].Feature)
	assert.Equal(t, "Feature 2", loaded[1].Feature)
}

func TestValidateEditContent_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	yamlContent := "{{{invalid yaml"
	err := validateEditContent(filename, []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "YAML parse error")
}

func TestValidateEditContent_EmptyContent(t *testing.T) {
	tmpDir := t.TempDir()
	filename := tmpDir + "/test-plan.yaml"

	err := validateEditContent(filename, []byte{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no valid test plans found")
}

func TestValidateEditContent_FileError(t *testing.T) {
	yamlContent := `feature: "Test"
type: "functional"
status: "planned"
scenarios:
  - "Scenario"
`
	err := validateEditContent("/nonexistent/dir/file.yaml", []byte(yamlContent))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")
}
