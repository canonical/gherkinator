package serve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gherkinator/internal/common"
)

func TestGenerateSphinxDocs_OrganisesByType(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Login Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
    When the user logs in
    Then the user sees the dashboard
---
feature: "Stress Test"
type: "performance"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Load test
    Given the system is running
    When 1000 users connect
    Then response time is under 500ms
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	plans, err := GenerateSphinxDocs(inputFile, docsDir, "", "")
	require.NoError(t, err)
	assert.Len(t, plans, 2)

	assert.FileExists(t, filepath.Join(docsDir, "functional", "login_feature.md"))
	assert.FileExists(t, filepath.Join(docsDir, "performance", "stress_test.md"))

	content, err := os.ReadFile(filepath.Join(docsDir, "functional", "login_feature.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Login Feature")
}

func TestGenerateSphinxDocs_EmptyFeatureName(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: ""
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    A scenario
    Given x
    When y
    Then z
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	plans, err := GenerateSphinxDocs(inputFile, docsDir, "", "")
	require.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.FileExists(t, filepath.Join(docsDir, "functional", "plan.md"))
}

func TestGenerateSphinxDocs_FileNotFound(t *testing.T) {
	_, err := GenerateSphinxDocs("/nonexistent.yaml", "/tmp", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestGenerateSphinxDocs_InvalidSchema(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Bad"
type: "invalid"
status: "planned"
risk: "stable"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	_, err := GenerateSphinxDocs(inputFile, tmpDir, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

func TestGenerateSphinxDocs_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "bad.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte("{{{bad"), 0644))

	_, err := GenerateSphinxDocs(inputFile, tmpDir, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode YAML")
}

func TestBuildTypeLandingPages_CreatesLandingPages(t *testing.T) {
	tmpDir := t.TempDir()
	grouped := map[string][]string{
		"functional":  {"login_feature", "signup_feature"},
		"performance": {"stress_test"},
	}

	err := BuildTypeLandingPages(tmpDir, grouped)
	require.NoError(t, err)

	// Functional landing page
	funcIndex := filepath.Join(tmpDir, "functional", "index.md")
	assert.FileExists(t, funcIndex)
	content, err := os.ReadFile(funcIndex)
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "# Functional")
	assert.Contains(t, s, "```{toctree}")
	assert.Contains(t, s, "login_feature")
	assert.Contains(t, s, "signup_feature")

	// Performance landing page
	perfIndex := filepath.Join(tmpDir, "performance", "index.md")
	assert.FileExists(t, perfIndex)
	content, err = os.ReadFile(perfIndex)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Performance")
	assert.Contains(t, string(content), "stress_test")
}

func TestBuildTypeLandingPages_SkipsEmptyTypes(t *testing.T) {
	tmpDir := t.TempDir()
	grouped := map[string][]string{
		"functional": {"login"},
	}

	err := BuildTypeLandingPages(tmpDir, grouped)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(tmpDir, "functional", "index.md"))
	// Types without plans should not have directories created
	assert.NoDirExists(t, filepath.Join(tmpDir, "security"))
	assert.NoDirExists(t, filepath.Join(tmpDir, "performance"))
}

func TestBuildTypeLandingPages_EmptyGrouped(t *testing.T) {
	tmpDir := t.TempDir()
	err := BuildTypeLandingPages(tmpDir, map[string][]string{})
	require.NoError(t, err)
}

func TestBuildTypeLandingPages_InvalidDir(t *testing.T) {
	grouped := map[string][]string{
		"functional": {"test"},
	}
	// Use a path where MkdirAll will fail (file exists as non-directory)
	tmpDir := t.TempDir()
	// Create a file where the directory should be
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "functional"), []byte("block"), 0644))

	err := BuildTypeLandingPages(tmpDir, grouped)
	assert.Error(t, err)
}

func TestBuildSphinxIndex_CreatesToctreeWithLandingPages(t *testing.T) {
	tmpDir := t.TempDir()
	plans := []common.TestPlan{
		{Feature: "Login Feature", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Stress Test", Type: "performance", Status: "implemented", Risk: "stable"},
		{Feature: "Auth Check", Type: "security", Status: "planned", Risk: "stable"},
	}

	err := BuildSphinxIndex(tmpDir, plans)
	require.NoError(t, err)

	// Root index should reference type landing pages
	indexPath := filepath.Join(tmpDir, "index.md")
	assert.FileExists(t, indexPath)
	content, err := os.ReadFile(indexPath)
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "# Test Plans")
	assert.Contains(t, s, "```{toctree}")
	assert.Contains(t, s, "functional/index")
	assert.Contains(t, s, "performance/index")
	assert.Contains(t, s, "security/index")

	// Test types should appear as level 2 headers
	assert.Contains(t, s, "## Functional")
	assert.Contains(t, s, "## Performance")
	assert.Contains(t, s, "## Security")

	// Feature names should appear as bullet points with links
	assert.Contains(t, s, "- [Login Feature](functional/login_feature.md)")
	assert.Contains(t, s, "- [Stress Test](performance/stress_test.md)")
	assert.Contains(t, s, "- [Auth Check](security/auth_check.md)")

	// Type landing pages should exist with their own toctrees
	funcContent, err := os.ReadFile(filepath.Join(tmpDir, "functional", "index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(funcContent), "login_feature")

	perfContent, err := os.ReadFile(filepath.Join(tmpDir, "performance", "index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(perfContent), "stress_test")
}

func TestBuildSphinxIndex_EmptyPlans(t *testing.T) {
	tmpDir := t.TempDir()

	err := BuildSphinxIndex(tmpDir, []common.TestPlan{})
	require.NoError(t, err)

	content, err := os.ReadFile(filepath.Join(tmpDir, "index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Test Plans")
}

func TestBuildSphinxIndex_NilPlans(t *testing.T) {
	tmpDir := t.TempDir()

	err := BuildSphinxIndex(tmpDir, nil)
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(tmpDir, "index.md"))
}

func TestBuildSphinxIndex_MultipleSameType(t *testing.T) {
	tmpDir := t.TempDir()
	plans := []common.TestPlan{
		{Feature: "Feature A", Type: "functional", Status: "planned", Risk: "stable"},
		{Feature: "Feature B", Type: "functional", Status: "implemented", Risk: "stable"},
	}

	err := BuildSphinxIndex(tmpDir, plans)
	require.NoError(t, err)

	// Root index references functional/index in toctree
	content, err := os.ReadFile(filepath.Join(tmpDir, "index.md"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "functional/index")
	assert.Contains(t, s, "## Functional")
	assert.Contains(t, s, "- [Feature A](functional/feature_a.md)")
	assert.Contains(t, s, "- [Feature B](functional/feature_b.md)")

	// Landing page has both files
	funcContent, err := os.ReadFile(filepath.Join(tmpDir, "functional", "index.md"))
	require.NoError(t, err)
	fs := string(funcContent)
	assert.Contains(t, fs, "feature_a")
	assert.Contains(t, fs, "feature_b")
}

func TestBuildSphinxIndex_InvalidDir(t *testing.T) {
	err := BuildSphinxIndex("/nonexistent/dir", []common.TestPlan{
		{Feature: "Test", Type: "functional", Status: "planned", Risk: "stable"},
	})
	assert.Error(t, err)
}

// helper to simulate a cloned canonical/sphinx-stack repository with a
// docs/ subdirectory that contains the template content directories and
// default index.rst that gherkinator prunes before generating docs.
func setupFakeSphinxStack(t *testing.T, cloneDir string) {
	t.Helper()
	docsDir := filepath.Join(cloneDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// Template content directories that PruneSphinxStackDefaults removes.
	for _, dir := range sphinxStackTemplateDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(docsDir, dir), 0755))
	}
	// Directories that must be preserved by pruning.
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "_dev"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "_templates"), 0755))

	// Default index.rst that is replaced by gherkinator's generated
	// index.md.  PruneSphinxStackDefaults removes this file.
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "index.rst"), []byte("Project\n=======\n"), 0644))

	confPy := `import datetime
import os
import textwrap

project = "Project"
author = "Canonical Ltd."

extensions = [
    "canonical_sphinx",
    "sphinx_rerediraffe",
    "sphinx_sitemap",
]

rediraffe_redirects = "redirects.txt"
rediraffe_dir_only = True

# disable_feedback_button = True

llms_txt_description = textwrap.dedent(
    """\
    This is the documentation for the Sphinx Stack, a template repository that helps you
    set up, build, and publish Sphinx documentation.
    """
)
`
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "conf.py"), []byte(confPy), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "Makefile"), []byte("make"), 0644))
}

func TestPrepareSphinxSite_EndToEnd(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Login Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
    When the user logs in
    Then the user sees the dashboard
---
feature: "Stress Test"
type: "performance"
status: "implemented"
risk: "stable"
scenarios:
  - |
    Load test
    Given the system is running
    When 1000 users connect
    Then response time is under 500ms
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	cloneDir := filepath.Join(tmpDir, "clone")
	setupFakeSphinxStack(t, cloneDir)
	docsDir := filepath.Join(cloneDir, "docs")

	err := PrepareSphinxSite(inputFile, cloneDir, "My Project", "", "")
	require.NoError(t, err)

	// conf.py should have the updated project name
	confContent, err := os.ReadFile(filepath.Join(docsDir, "conf.py"))
	require.NoError(t, err)
	confStr := string(confContent)
	assert.Contains(t, confStr, `project = "My Project"`)
	assert.NotContains(t, confStr, `project = "Project"`)
	// Feedback button should be disabled (uncommented)
	assert.Contains(t, confStr, "disable_feedback_button = True")
	assert.NotContains(t, confStr, "# disable_feedback_button = True")
	// Rediraffe should be removed
	assert.NotContains(t, confStr, "rediraffe_redirects")
	assert.NotContains(t, confStr, "rediraffe_dir_only")
	assert.NotContains(t, confStr, "sphinx_rerediraffe")
	// llms_txt_description should be replaced with the project-specific
	// gherkinator description and not contain the sphinx-stack default.
	assert.Contains(t, confStr, "Gherkin-based test plans for My Project")
	assert.NotContains(t, confStr, "Sphinx Stack, a template repository")

	// conf.py and Makefile should remain
	assert.FileExists(t, filepath.Join(docsDir, "conf.py"))
	assert.FileExists(t, filepath.Join(docsDir, "Makefile"))

	// sphinx-stack template content directories and default index.rst
	// should have been pruned before generation.
	assert.NoDirExists(t, filepath.Join(docsDir, "contribute"))
	assert.NoDirExists(t, filepath.Join(docsDir, "explanation"))
	assert.NoDirExists(t, filepath.Join(docsDir, "how-to"))
	assert.NoDirExists(t, filepath.Join(docsDir, "reference"))
	assert.NoDirExists(t, filepath.Join(docsDir, "release-notes"))
	assert.NoDirExists(t, filepath.Join(docsDir, "tutorials"))
	assert.NoFileExists(t, filepath.Join(docsDir, "index.rst"))
	// _dev/ and _templates/ must be preserved.
	assert.DirExists(t, filepath.Join(docsDir, "_dev"))
	assert.DirExists(t, filepath.Join(docsDir, "_templates"))

	// Generated markdown in type subdirs
	assert.FileExists(t, filepath.Join(docsDir, "functional", "login_feature.md"))
	assert.FileExists(t, filepath.Join(docsDir, "performance", "stress_test.md"))

	// Type landing pages
	assert.FileExists(t, filepath.Join(docsDir, "functional", "index.md"))
	funcContent, err := os.ReadFile(filepath.Join(docsDir, "functional", "index.md"))
	require.NoError(t, err)
	assert.Contains(t, string(funcContent), "# Functional")
	assert.Contains(t, string(funcContent), "login_feature")

	assert.FileExists(t, filepath.Join(docsDir, "performance", "index.md"))

	// Root index has toctree referencing type landing pages and
	// level 2 headers with feature bullet points
	content, err := os.ReadFile(filepath.Join(docsDir, "index.md"))
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, "# Test Plans")
	assert.Contains(t, s, "functional/index")
	assert.Contains(t, s, "performance/index")
	assert.Contains(t, s, "## Functional")
	assert.Contains(t, s, "## Performance")
	assert.Contains(t, s, "- [Login Feature](functional/login_feature.md)")
	assert.Contains(t, s, "- [Stress Test](performance/stress_test.md)")
}

func TestPrepareSphinxSite_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cloneDir := filepath.Join(tmpDir, "clone")
	setupFakeSphinxStack(t, cloneDir)

	inputFile := filepath.Join(tmpDir, "bad.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte("{{{bad"), 0644))

	err := PrepareSphinxSite(inputFile, cloneDir, "test", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate sphinx docs")
}

func TestPrepareSphinxSite_NonexistentYAML(t *testing.T) {
	tmpDir := t.TempDir()
	cloneDir := filepath.Join(tmpDir, "clone")
	setupFakeSphinxStack(t, cloneDir)

	err := PrepareSphinxSite("/nonexistent.yaml", cloneDir, "test", "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate sphinx docs")
}

func TestPrepareSphinxSite_NonexistentDocsDir(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(""), 0644))

	err := PrepareSphinxSite(inputFile, filepath.Join(tmpDir, "nonexistent"), "test", "", "")
	assert.Error(t, err)
}

func TestGenerateSphinxDocs_RiskFilter(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Edge Feature"
type: "functional"
status: "planned"
risk: "edge"
scenarios:
  - "Edge scenario"
---
feature: "Beta Feature"
type: "functional"
status: "planned"
risk: "beta"
scenarios:
  - "Beta scenario"
---
feature: "Stable Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Stable scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// Test risk filter beta (should include edge and beta, but not stable)
	plans, err := GenerateSphinxDocs(inputFile, docsDir, "beta", "")
	require.NoError(t, err)
	assert.Len(t, plans, 2)
	assert.FileExists(t, filepath.Join(docsDir, "functional", "edge_feature.md"))
	assert.FileExists(t, filepath.Join(docsDir, "functional", "beta_feature.md"))
	assert.NoFileExists(t, filepath.Join(docsDir, "security", "stable_feature.md"))
}

func TestGenerateSphinxDocs_RiskFilterEdge(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Edge Feature"
type: "functional"
status: "planned"
risk: "edge"
scenarios:
  - "Edge scenario"
---
feature: "Beta Feature"
type: "functional"
status: "planned"
risk: "beta"
scenarios:
  - "Beta scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// Test risk filter edge (should only include edge)
	plans, err := GenerateSphinxDocs(inputFile, docsDir, "edge", "")
	require.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.FileExists(t, filepath.Join(docsDir, "functional", "edge_feature.md"))
	assert.NoFileExists(t, filepath.Join(docsDir, "functional", "beta_feature.md"))
}

func TestUpdateConfPy_UpdatesProjectName(t *testing.T) {
	tmpDir := t.TempDir()
	confPy := `import textwrap

project = "Project"
author = "Canonical Ltd."
# disable_feedback_button = True
rediraffe_redirects = "redirects.txt"
rediraffe_dir_only = True
extensions = [
    "canonical_sphinx",
    "sphinx_rerediraffe",
    "sphinx_sitemap",
]
llms_txt_description = textwrap.dedent(
    """\
    This is the documentation for the Sphinx Stack, a template repository that helps you
    set up, build, and publish Sphinx documentation.
    """
)
`
	confPath := filepath.Join(tmpDir, "conf.py")
	require.NoError(t, os.WriteFile(confPath, []byte(confPy), 0644))

	err := UpdateConfPy(confPath, "charmed-hpc")
	require.NoError(t, err)

	content, err := os.ReadFile(confPath)
	require.NoError(t, err)
	s := string(content)
	assert.Contains(t, s, `project = "charmed-hpc"`)
	assert.NotContains(t, s, `project = "Project"`)
	assert.Contains(t, s, `author = "Canonical Ltd."`)
	// Feedback button should be uncommented
	assert.Contains(t, s, "disable_feedback_button = True")
	assert.NotContains(t, s, "# disable_feedback_button = True")
	// Rediraffe should be removed
	assert.NotContains(t, s, "rediraffe_redirects")
	assert.NotContains(t, s, "rediraffe_dir_only")
	assert.NotContains(t, s, "sphinx_rerediraffe")
	// llms_txt_description should be replaced with the project-specific text
	assert.Contains(t, s, "Gherkin-based test plans for charmed-hpc")
	assert.NotContains(t, s, "Sphinx Stack, a template repository")
	// Other extensions should remain
	assert.Contains(t, s, "canonical_sphinx")
	assert.Contains(t, s, "sphinx_sitemap")
}

func TestUpdateConfPy_FileNotFound(t *testing.T) {
	err := UpdateConfPy("/nonexistent/conf.py", "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read")
}

func TestUpdateConfPy_MissingProjectVariable(t *testing.T) {
	tmpDir := t.TempDir()
	confPath := filepath.Join(tmpDir, "conf.py")
	require.NoError(t, os.WriteFile(confPath, []byte("author = \"Test\"\n"), 0644))

	err := UpdateConfPy(confPath, "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find 'project' variable")
}

func TestUpdateConfPy_SpecialCharsInName(t *testing.T) {
	tmpDir := t.TempDir()
	confPy := "project = \"Old Name\"\n"
	confPath := filepath.Join(tmpDir, "conf.py")
	require.NoError(t, os.WriteFile(confPath, []byte(confPy), 0644))

	err := UpdateConfPy(confPath, `My "Special" Project`)
	require.NoError(t, err)

	content, err := os.ReadFile(confPath)
	require.NoError(t, err)
	// Go's %q properly escapes quotes
	assert.Contains(t, string(content), `project = "My \"Special\" Project"`)
}

func TestPruneSphinxStackDefaults_RemovesTemplateDirs(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// Populate every directory the function is expected to remove.
	for _, dir := range sphinxStackTemplateDirs {
		require.NoError(t, os.MkdirAll(filepath.Join(docsDir, dir), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(docsDir, dir, "page.md"), []byte("x"), 0644))
	}
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "index.rst"), []byte("Project\n=======\n"), 0644))

	// Populate the directories that must be preserved.
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "_dev"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "_dev", "keep.txt"), []byte("keep"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(docsDir, "_templates"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "_templates", "keep.html"), []byte("keep"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "conf.py"), []byte("p"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(docsDir, "Makefile"), []byte("m"), 0644))

	require.NoError(t, PruneSphinxStackDefaults(docsDir))

	// Template content directories and the default index.rst are gone.
	for _, dir := range sphinxStackTemplateDirs {
		assert.NoDirExists(t, filepath.Join(docsDir, dir), "expected %s to be removed", dir)
	}
	assert.NoFileExists(t, filepath.Join(docsDir, "index.rst"))

	// _dev/, _templates/, conf.py, and Makefile are preserved.
	assert.DirExists(t, filepath.Join(docsDir, "_dev"))
	assert.FileExists(t, filepath.Join(docsDir, "_dev", "keep.txt"))
	assert.DirExists(t, filepath.Join(docsDir, "_templates"))
	assert.FileExists(t, filepath.Join(docsDir, "_templates", "keep.html"))
	assert.FileExists(t, filepath.Join(docsDir, "conf.py"))
	assert.FileExists(t, filepath.Join(docsDir, "Makefile"))
}

func TestPruneSphinxStackDefaults_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))
	// A fresh clone with no template content; the function must not
	// error when there is nothing to remove.
	require.NoError(t, PruneSphinxStackDefaults(docsDir))
	require.NoError(t, PruneSphinxStackDefaults(docsDir))
}

func TestPruneSphinxStackDefaults_DocsDirDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	docsDir := filepath.Join(tmpDir, "does-not-exist")
	// Pruning against a missing docs directory should not error: each
	// RemoveAll on a missing path is a no-op and index.rst is absent.
	require.NoError(t, PruneSphinxStackDefaults(docsDir))
}

func TestCleanGeneratedDocs_RemovesTypeDirs(t *testing.T) {
	tmpDir := t.TempDir()
	for _, sub := range []string{"functional", "performance", "security"} {
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, sub), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, sub, "feature.md"), []byte("x"), 0644))
	}
	// conf.py and Makefile should remain untouched
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "conf.py"), []byte("p"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte("m"), 0644))

	err := CleanGeneratedDocs(tmpDir)
	require.NoError(t, err)

	assert.NoDirExists(t, filepath.Join(tmpDir, "functional"))
	assert.NoDirExists(t, filepath.Join(tmpDir, "performance"))
	assert.NoDirExists(t, filepath.Join(tmpDir, "security"))
	assert.FileExists(t, filepath.Join(tmpDir, "conf.py"))
	assert.FileExists(t, filepath.Join(tmpDir, "Makefile"))
}

func TestCleanGeneratedDocs_NoTypeDirs(t *testing.T) {
	tmpDir := t.TempDir()
	err := CleanGeneratedDocs(tmpDir)
	assert.NoError(t, err)
}

func TestCleanGeneratedDocs_UnknownTypePreserved(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "unrelated"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "unrelated", "file.md"), []byte("x"), 0644))

	err := CleanGeneratedDocs(tmpDir)
	require.NoError(t, err)
	// Directories not in validTestTypes must be preserved
	assert.DirExists(t, filepath.Join(tmpDir, "unrelated"))
}

func TestGenerateSphinxDocs_StatusFilter(t *testing.T) {
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

	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	// status=planned: only the planned feature renders
	plans, err := GenerateSphinxDocs(inputFile, docsDir, "", "planned")
	require.NoError(t, err)
	assert.Len(t, plans, 1)
	assert.FileExists(t, filepath.Join(docsDir, "functional", "planned_feature.md"))
	assert.NoFileExists(t, filepath.Join(docsDir, "security", "implemented_feature.md"))
	assert.NoFileExists(t, filepath.Join(docsDir, "solution", "deprecated_feature.md"))
}

func TestGenerateSphinxDocs_BothFilters_ImplementedAndCandidate(t *testing.T) {
	// --status=implemented --risk=candidate: only "implemented" plans whose
	// risk is edge, beta, or candidate.
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
  - "Planned beta scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	plans, err := GenerateSphinxDocs(inputFile, docsDir, "candidate", "implemented")
	require.NoError(t, err)
	assert.Len(t, plans, 2)
	assert.FileExists(t, filepath.Join(docsDir, "functional", "implemented_edge.md"))
	assert.FileExists(t, filepath.Join(docsDir, "security", "implemented_candidate.md"))
	assert.NoFileExists(t, filepath.Join(docsDir, "solution", "implemented_stable.md"))
	assert.NoFileExists(t, filepath.Join(docsDir, "functional", "planned_beta.md"))
}

func TestGenerateSphinxDocs_BothFilters_ImplementedAndStable(t *testing.T) {
	// --risk=stable matches every risk level, so the intersection with
	// --status=implemented should yield only implemented plans.
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
  - "Planned stable scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	docsDir := filepath.Join(tmpDir, "docs")
	require.NoError(t, os.MkdirAll(docsDir, 0755))

	plans, err := GenerateSphinxDocs(inputFile, docsDir, "stable", "implemented")
	require.NoError(t, err)
	assert.Len(t, plans, 2)
	assert.FileExists(t, filepath.Join(docsDir, "functional", "implemented_edge.md"))
	assert.FileExists(t, filepath.Join(docsDir, "security", "implemented_stable.md"))
	assert.NoFileExists(t, filepath.Join(docsDir, "solution", "planned_stable.md"))
}
