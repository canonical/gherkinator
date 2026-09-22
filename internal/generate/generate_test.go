package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/gherkinator/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessFile_GherkinFormat(t *testing.T) {
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
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "gh", outputDir, common.FilterOptions{})
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
	err = ProcessFile(inputFile, "md", outputDir, common.FilterOptions{})
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
risk: "stable"
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
risk: "stable"
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
	err = ProcessFile(inputFile, "md", outputDir, common.FilterOptions{})
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "feature_one.md"))
	assert.FileExists(t, filepath.Join(outputDir, "feature_two.md"))
}

func TestProcessFile_InvalidSchema(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `feature: "Bad Feature"
type: "invalid_type"
status: "planned"
risk: "stable"
scenarios:
  - "test"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "gh", outputDir, common.FilterOptions{})
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
	err = ProcessFile(inputFile, "gh", outputDir, common.FilterOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode YAML")
}

func TestProcessFile_FileNotFound(t *testing.T) {
	err := ProcessFile("/nonexistent/file.yaml", "gh", "/tmp/out", common.FilterOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestProcessFile_UnsupportedFormat(t *testing.T) {
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
	err = ProcessFile(inputFile, "xml", outputDir, common.FilterOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestProcessFile_EmptyFeatureName(t *testing.T) {
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
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	err = ProcessFile(inputFile, "md", outputDir, common.FilterOptions{})
	assert.NoError(t, err)

	// Should use plan_1 as fallback filename
	assert.FileExists(t, filepath.Join(outputDir, "plan_1.md"))
}

func TestProcessFile_RiskFilter(t *testing.T) {
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
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")

	// Test --risk=beta (should include edge and beta, but not stable)
	err = ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Risk: "beta"})
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "edge_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "beta_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "stable_feature.md"))
}

func TestProcessFile_RiskFilterEdge(t *testing.T) {
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
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")

	// Test --risk=edge (should only include edge)
	err = ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Risk: "edge"})
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "edge_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "beta_feature.md"))
}

func TestProcessFile_RiskFilterStable(t *testing.T) {
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
feature: "Candidate Feature"
type: "security"
status: "planned"
risk: "candidate"
scenarios:
  - "Candidate scenario"
---
feature: "Stable Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Stable scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	err := os.WriteFile(inputFile, []byte(yamlContent), 0644)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")

	// Test --risk=stable (should include all)
	err = ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Risk: "stable"})
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "edge_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "beta_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "candidate_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "stable_feature.md"))
}

func TestProcessFile_StatusFilter_Planned(t *testing.T) {
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
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Status: "planned"})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "planned_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "implemented_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "deprecated_feature.md"))
}

func TestProcessFile_StatusFilter_Implemented(t *testing.T) {
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
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Status: "implemented"})
	require.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(outputDir, "planned_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "implemented_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "deprecated_feature.md"))
}

func TestProcessFile_StatusFilter_Deprecated(t *testing.T) {
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
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Status: "deprecated"})
	require.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(outputDir, "planned_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "implemented_feature.md"))
	assert.FileExists(t, filepath.Join(outputDir, "deprecated_feature.md"))
}

func TestProcessFile_StatusFilter_NoMatches(t *testing.T) {
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
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	// No plan is deprecated — should produce no output files.
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Status: "deprecated"})
	require.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(outputDir, "planned_feature.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "implemented_feature.md"))
}

func TestProcessFile_BothFilters_ImplementedAndCandidate(t *testing.T) {
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
feature: "Implemented Beta"
type: "functional"
status: "implemented"
risk: "beta"
scenarios:
  - "Beta scenario"
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

	outputDir := filepath.Join(tmpDir, "output")
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Risk: "candidate", Status: "implemented"})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "implemented_edge.md"))
	assert.FileExists(t, filepath.Join(outputDir, "implemented_beta.md"))
	assert.FileExists(t, filepath.Join(outputDir, "implemented_candidate.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "implemented_stable.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "planned_beta.md"))
}

func TestProcessFile_BothFilters_PlannedAndBeta(t *testing.T) {
	// --status=planned --risk=beta: only "planned" plans whose risk is
	// edge or beta.
	tmpDir := t.TempDir()
	yamlContent := `feature: "Planned Edge"
type: "functional"
status: "planned"
risk: "edge"
scenarios:
  - "Planned edge scenario"
---
feature: "Planned Beta"
type: "functional"
status: "planned"
risk: "beta"
scenarios:
  - "Planned beta scenario"
---
feature: "Planned Candidate"
type: "performance"
status: "planned"
risk: "candidate"
scenarios:
  - "Planned candidate scenario"
---
feature: "Planned Stable"
type: "security"
status: "planned"
risk: "stable"
scenarios:
  - "Planned stable scenario"
---
feature: "Implemented Beta"
type: "functional"
status: "implemented"
risk: "beta"
scenarios:
  - "Implemented beta scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Risk: "beta", Status: "planned"})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "planned_edge.md"))
	assert.FileExists(t, filepath.Join(outputDir, "planned_beta.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "planned_candidate.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "planned_stable.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "implemented_beta.md"))
}

func TestProcessFile_TagFilter(t *testing.T) {
	// --tag renders only plans carrying the requested tag; untagged
	// plans are excluded while the tag filter is active.
	tmpDir := t.TempDir()
	yamlContent := `feature: "Multi Node"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - multi-node
scenarios:
  - "Multi node scenario"
---
feature: "Single Node"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - single-node
scenarios:
  - "Single node scenario"
---
feature: "Untagged"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "Untagged scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Tags: []string{"multi-node"}})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "multi_node.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "single_node.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "untagged.md"))
}

func TestProcessFile_TagFilterUnion(t *testing.T) {
	// Multiple requested tags use union semantics: a plan matching any
	// requested tag renders.
	tmpDir := t.TempDir()
	yamlContent := `feature: "Multi Node"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - multi-node
scenarios:
  - "Multi node scenario"
---
feature: "Minimal"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - minimal
scenarios:
  - "Minimal scenario"
---
feature: "Single Node"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - single-node
scenarios:
  - "Single node scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Tags: []string{"multi-node", "minimal"}})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "multi_node.md"))
	assert.FileExists(t, filepath.Join(outputDir, "minimal.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "single_node.md"))
}

func TestProcessFile_RiskAndTagFilterIntersection(t *testing.T) {
	// --risk=edge --tag multi-node (the issue #6 example): only plans
	// that are edge risk AND tagged multi-node render.
	tmpDir := t.TempDir()
	yamlContent := `feature: "Edge Multi Node"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - multi-node
scenarios:
  - "Edge multi node scenario"
---
feature: "Stable Multi Node"
type: "functional"
status: "planned"
risk: "stable"
tags:
  - multi-node
scenarios:
  - "Stable multi node scenario"
---
feature: "Edge Single Node"
type: "functional"
status: "planned"
risk: "edge"
tags:
  - single-node
scenarios:
  - "Edge single node scenario"
`
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(yamlContent), 0644))

	outputDir := filepath.Join(tmpDir, "output")
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Risk: "edge", Tags: []string{"multi-node"}})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "edge_multi_node.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "stable_multi_node.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "edge_single_node.md"))
}

func TestProcessFile_BothFilters_ImplementedAndStable(t *testing.T) {
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

	outputDir := filepath.Join(tmpDir, "output")
	err := ProcessFile(inputFile, "md", outputDir, common.FilterOptions{Risk: "stable", Status: "implemented"})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(outputDir, "implemented_edge.md"))
	assert.FileExists(t, filepath.Join(outputDir, "implemented_stable.md"))
	assert.NoFileExists(t, filepath.Join(outputDir, "planned_stable.md"))
}
