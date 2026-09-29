package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeYAML(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(body), 0644))
	return p
}

func TestSafeFeatureName_ReplacesSpacesWithUnderscores(t *testing.T) {
	assert.Equal(t, "submit_mpi_job", SafeFeatureName(TestPlan{Feature: "Submit MPI job"}, "fallback"))
}

func TestSafeFeatureName_Lowercases(t *testing.T) {
	assert.Equal(t, "submit_mpi_job", SafeFeatureName(TestPlan{Feature: "Submit MPI Job"}, "fallback"))
}

func TestSafeFeatureName_MultipleSpacesCollapseToUnderscores(t *testing.T) {
	assert.Equal(t, "a___b", SafeFeatureName(TestPlan{Feature: "A   B"}, "fallback"))
}

func TestSafeFeatureName_EmptyUsesFallback(t *testing.T) {
	assert.Equal(t, "plan_3", SafeFeatureName(TestPlan{Feature: ""}, "plan_3"))
}

func TestRenderPlans_GherkinFormat(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "login_feature.feature", files[0].Filename)
	assert.Contains(t, files[0].Content, "Feature: Login Feature")
	assert.Contains(t, files[0].Content, "Scenario: User logs in")
}

func TestRenderPlans_MarkdownFormat(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	files, err := RenderPlans([]string{p}, "md", FilterOptions{})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "login_feature.md", files[0].Filename)
	assert.Contains(t, files[0].Content, "# Login Feature")
}

func TestRenderPlans_MultiDocumentYAML(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: "Feature One"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    Scenario One
    Given x
---
feature: "Feature Two"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Scenario Two
    Given y
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{})
	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, "feature_one.feature", files[0].Filename)
	assert.Equal(t, "feature_two.feature", files[1].Filename)
}

func TestRenderPlans_MultipleInputFiles(t *testing.T) {
	tmpDir := t.TempDir()
	p1 := writeYAML(t, tmpDir, "a.yaml", `feature: "Alpha"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    A scenario
    Given x
`)
	p2 := writeYAML(t, tmpDir, "b.yaml", `feature: "Beta"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - |
    B scenario
    Given y
`)

	files, err := RenderPlans([]string{p1, p2}, "gh", FilterOptions{})
	require.NoError(t, err)
	require.Len(t, files, 2)
	names := []string{files[0].Filename, files[1].Filename}
	assert.Contains(t, names, "alpha.feature")
	assert.Contains(t, names, "beta.feature")
}

func TestRenderPlans_EmptyFeatureNameFallsBack(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: ""
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    A scenario
    Given x
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "plan_1.feature", files[0].Filename)
}

func TestRenderPlans_EmptyFeatureNameMarkdownUsesPlanFallback(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: ""
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    A scenario
    Given x
`)

	files, err := RenderPlans([]string{p}, "md", FilterOptions{})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "plan_1.md", files[0].Filename)
}

func TestRenderPlans_EmptyFeatureNameMultiDocFallbacksAreDistinct(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: ""
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    Scenario 1
    Given x
---
feature: ""
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Scenario 2
    Given y
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{})
	require.NoError(t, err)
	require.Len(t, files, 2)
	assert.Equal(t, "plan_1.feature", files[0].Filename)
	assert.Equal(t, "plan_2.feature", files[1].Filename)
}

func TestRenderPlans_StatusFilter(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: "Planned"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "Planned scenario"
---
feature: "Implemented"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Implemented scenario"
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{Status: "planned"})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "planned.feature", files[0].Filename)
}

func TestRenderPlans_RiskFilter(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: "Edge"
type: "functional"
status: "planned"
risk: "edge"
scenarios:
  - "Edge scenario"
---
feature: "Stable"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Stable scenario"
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{Risk: "beta"})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "edge.feature", files[0].Filename)
}

func TestRenderPlans_TagFilter(t *testing.T) {
	// With a tag filter active, only plans carrying a matching tag
	// render and the custom tags appear in the Gherkin output.
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: "Tagged"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - multi-node
scenarios:
  - "Tagged scenario"
---
feature: "Untagged"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Untagged scenario"
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{Tags: []string{"multi-node"}})
	require.NoError(t, err)
	require.Len(t, files, 1)
	assert.Equal(t, "tagged.feature", files[0].Filename)
	assert.Contains(t, files[0].Content, "@functional @edge @planned @multi-node")
}

func TestRenderPlans_FileNotFound(t *testing.T) {
	files, err := RenderPlans([]string{"/nonexistent/file.yaml"}, "gh", FilterOptions{})
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestRenderPlans_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "bad.yaml", `{{{not valid yaml`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{})
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to decode YAML")
}

func TestRenderPlans_InvalidSchema(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "bad.yaml", `feature: "Bad"
type: "invalid_type"
status: "planned"
risk: "stable"
scenarios:
  - "scenario"
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{})
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "validation error")
}

func TestRenderPlans_InvalidSchemaInMultiDoc(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "mixed.yaml", `feature: "Good"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "scenario"
---
feature: "Bad"
type: "invalid_type"
status: "planned"
risk: "stable"
scenarios:
  - "scenario"
`)

	files, err := RenderPlans([]string{p}, "gh", FilterOptions{})
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "document 2")
}

func TestRenderPlans_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	p := writeYAML(t, tmpDir, "p.yaml", `feature: "Test"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "scenario"
`)

	files, err := RenderPlans([]string{p}, "xml", FilterOptions{})
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestExtension(t *testing.T) {
	assert.Equal(t, ".feature", extension("gh"))
	assert.Equal(t, ".md", extension("md"))
	assert.Equal(t, "", extension("bogus"))
	assert.Equal(t, "", extension(""))
}

func TestFallbackName(t *testing.T) {
	assert.Equal(t, "plan_1", fallbackName("gh", 1))
	assert.Equal(t, "plan_42", fallbackName("md", 42))
}
