package edit

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/gherkinator/internal/common"

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

// captureStderr collects everything written to os.Stderr while fn runs. The
// swap is safe here because the tests are single-goroutine.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	fn()

	require.NoError(t, w.Close())
	os.Stderr = oldStderr
	output, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(output)
}

func TestResolveEditor_PrefersVISUALOverEDITOR(t *testing.T) {
	t.Setenv("VISUAL", "cat")
	t.Setenv("EDITOR", "true")

	editor, err := resolveEditor()
	require.NoError(t, err)
	assert.Equal(t, "cat", editor)
}

func TestResolveEditor_InaccessibleVISUALFallsBackToEDITOR(t *testing.T) {
	t.Setenv("VISUAL", "/nonexistent/vim")
	t.Setenv("EDITOR", "cat")

	var editor string
	stderr := captureStderr(t, func() {
		var err error
		editor, err = resolveEditor()
		require.NoError(t, err)
	})

	// The inaccessible $VISUAL is skipped with a warning and $EDITOR wins.
	assert.Equal(t, "cat", editor)
	assert.Contains(t, stderr, "warning: editor \"/nonexistent/vim\" from $VISUAL is not accessible")
}

func TestResolveEditor_WhitespaceOnlyEditorIgnored(t *testing.T) {
	t.Setenv("VISUAL", "   ")
	t.Setenv("EDITOR", "cat")

	// A whitespace-only value must not panic on strings.Fields and should
	// be treated as unset.
	editor, err := resolveEditor()
	require.NoError(t, err)
	assert.Equal(t, "cat", editor)
}

func TestResolveEditor_FallsBackToDefaultEditorList(t *testing.T) {
	// Provide a fake nano so the fallback list resolves inside the test,
	// mirroring the bundled editor inside the strictly confined snap.
	binDir := t.TempDir()
	nanoPath := filepath.Join(binDir, "nano")
	require.NoError(t, os.WriteFile(nanoPath, []byte("#!/bin/sh\nexit 0\n"), 0o755))
	t.Setenv("PATH", binDir)

	t.Setenv("VISUAL", "/nonexistent/vim")
	t.Setenv("EDITOR", "/nonexistent/nano")

	var editor string
	stderr := captureStderr(t, func() {
		var err error
		editor, err = resolveEditor()
		require.NoError(t, err)
	})

	assert.Equal(t, "nano", editor)
	assert.Contains(t, stderr, "$VISUAL is not accessible")
	assert.Contains(t, stderr, "$EDITOR is not accessible")
}

func TestResolveEditor_NoEditorAvailable(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "")

	path := os.Getenv("PATH")
	require.NoError(t, os.Setenv("PATH", "/nonexistent"))
	t.Cleanup(func() {
		require.NoError(t, os.Setenv("PATH", path))
	})

	_, err := resolveEditor()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no text editor found")
}

func TestTextEditor_InaccessibleVISUALFallsBack(t *testing.T) {
	// The fake nano echoes the file it is handed, exercising the full
	// fallback path from an inaccessible $VISUAL through to a valid editor.
	// /bin/cat is absolute because PATH is restricted to binDir.
	binDir := t.TempDir()
	nanoPath := filepath.Join(binDir, "nano")
	require.NoError(t, os.WriteFile(nanoPath, []byte("#!/bin/sh\n/bin/cat \"$1\"\n"), 0o755))
	t.Setenv("PATH", binDir)
	t.Setenv("VISUAL", "/nonexistent/vim")
	t.Setenv("EDITOR", "")

	content := []byte("test content")
	var edited []byte
	stderr := captureStderr(t, func() {
		var err error
		edited, err = TextEditor(content)
		require.NoError(t, err)
	})

	assert.Equal(t, content, edited)
	assert.Contains(t, stderr, "falling back to the default editor")
}

func TestTextEditor_EditorExitsWithError(t *testing.T) {
	binDir := t.TempDir()
	editorPath := filepath.Join(binDir, "failing-editor")
	require.NoError(t, os.WriteFile(editorPath, []byte("#!/bin/sh\nexit 1\n"), 0o755))
	t.Setenv("PATH", binDir)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "failing-editor")

	// An editor that launches but exits non-zero is a user abort, not a
	// resolution problem; it must not fall back to another editor.
	_, err := TextEditor([]byte("test content"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "editor exited with error")
}

func TestTextEditor_CreateTempFails(t *testing.T) {
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "cat")
	t.Setenv("TMPDIR", "/nonexistent-tmpdir-for-gherkinator-tests")

	// With a broken TMPDIR the scratch file cannot be created and the
	// failure must surface instead of silently losing the edit.
	_, err := TextEditor([]byte("test content"))
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "editor exited with error")
}

func TestTextEditor_EditorDeletesTempFile(t *testing.T) {
	binDir := t.TempDir()
	editorPath := filepath.Join(binDir, "deleting-editor")
	require.NoError(t, os.WriteFile(editorPath, []byte("#!/bin/sh\n/bin/rm \"$1\"\n"), 0o755))
	t.Setenv("PATH", binDir)
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", "deleting-editor")

	// If the scratch file disappears, reading back the edit must fail
	// rather than return empty content.
	_, err := TextEditor([]byte("test content"))
	require.Error(t, err)
}

func TestEditHelpTemplate_ContainsSchemaFields(t *testing.T) {
	template := EditHelpTemplate()
	assert.Contains(t, template, "feature:")
	assert.Contains(t, template, "type:")
	assert.Contains(t, template, "status:")
	assert.Contains(t, template, "risk:")
	assert.Contains(t, template, "scenarios:")
	assert.Contains(t, template, "examples:")
	assert.Contains(t, template, "tags:")
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
