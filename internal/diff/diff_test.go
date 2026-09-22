package diff

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/canonical/gherkinator/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeYAML(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(body), 0644))
	return p
}

func writeFeature(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(body), 0644))
}

func TestRun_CleanExitNoDiffs(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
    When the user logs in
    Then the user sees the dashboard
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	expected := common.GenerateGherkin(common.TestPlan{
		Feature: "Login Feature",
		Type:    "functional",
		Status:  "implemented",
		Risk:    "stable",
		Scenarios: []string{
			"User logs in\nGiven a user exists\nWhen the user logs in\nThen the user sees the dashboard",
		},
	})
	writeFeature(t, featuresDir, "login_feature.feature", expected)

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitClean, code)
	assert.Empty(t, stderr.String())
}

func TestRun_DifferingClassification(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "login_feature.feature", "Feature: Login Feature\nOLD STUFF\n")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "differs from feature file(s)")
	assert.Contains(t, stderr.String(), `"login_feature.feature"`)
}

func TestRun_MissingClassification(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "is missing")
	assert.Contains(t, stderr.String(), `"Login Feature"`)
}

func TestRun_OrphanClassification(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	expected := common.GenerateGherkin(common.TestPlan{
		Feature: "Login Feature",
		Type:    "functional",
		Status:  "implemented",
		Risk:    "stable",
		Scenarios: []string{
			"User logs in\nGiven a user exists",
		},
	})
	writeFeature(t, featuresDir, "login_feature.feature", expected)
	writeFeature(t, featuresDir, "orphan.feature", "Feature: Orphan\n")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "matches no test")
	assert.Contains(t, stderr.String(), `"orphan.feature"`)
}

func TestRun_AllThreeClassifications(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Existing Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    A scenario
    Given x
---
feature: "Missing Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    A scenario
    Given y
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "existing_feature.feature", "DIFFERENT CONTENT")
	writeFeature(t, featuresDir, "orphan.feature", "Feature: Orphan\n")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "differs")
	assert.Contains(t, stderr.String(), "missing")
	assert.Contains(t, stderr.String(), "matches no test")
}

func TestRun_TwoMissingPlansExercisesSort(t *testing.T) {
	tmpDir := t.TempDir()
	// Two unrelated plans → both files missing → exercises the
	// non-empty sort lambda in findMissing.
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Zeta Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "z scenario"
---
feature: "Alpha Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "a scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	out := stderr.String()
	assert.Contains(t, out, "are missing")
	// Output must be sorted alphabetically: Alpha before Zeta.
	alphaIdx := strings.Index(out, "Alpha Feature")
	zetaIdx := strings.Index(out, "Zeta Feature")
	assert.NotEqual(t, -1, alphaIdx)
	assert.NotEqual(t, -1, zetaIdx)
	assert.Less(t, alphaIdx, zetaIdx)
}

func TestRun_TwoDifferingPlansExercisesSort(t *testing.T) {
	tmpDir := t.TempDir()
	// Two plans that both differ from their on-disk files →
	// exercises the sort lambda in findDiffering.
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Zeta Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "z scenario"
---
feature: "Alpha Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "a scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "zeta_feature.feature", "ZETA OLD")
	writeFeature(t, featuresDir, "alpha_feature.feature", "ALPHA OLD")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "differ from")
}

func TestRun_DifferingExcludedFromOrphan(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Touched Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    A scenario
    Given x
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "touched_feature.feature", "FEATURE THAT DIFFERS")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.NotContains(t, stderr.String(), "matches no test")
	assert.Contains(t, stderr.String(), "differs")
}

func TestRun_InvalidSchemaYieldsPlanError(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Bad"
type: "invalid_type"
status: "planned"
risk: "stable"
scenarios:
  - "scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitPlanError, code)
	assert.Contains(t, stderr.String(), "validation error")
}

func TestRun_InvalidYAMLYieldsPlanError(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `{{{not valid yaml`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitPlanError, code)
	assert.Contains(t, stderr.String(), "decode YAML")
}

func TestRun_DuplicateFeatureYieldsPlanError(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Same Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario one"
---
feature: "Same Feature"
type: "security"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario two"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitPlanError, code)
	assert.Contains(t, stderr.String(), "two or more plans produce")
	assert.Contains(t, stderr.String(), "same_feature.feature")
}

func TestRun_FeaturesDirectoryMissingYieldsPlanError(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Foo"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)

	var stderr bytes.Buffer
	code := Run([]string{planPath}, filepath.Join(tmpDir, "nonexistent"), common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitPlanError, code)
	assert.Contains(t, stderr.String(), "features directory")
}

func TestRun_FeaturesPathNotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Foo"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)

	notADir := filepath.Join(tmpDir, "file.yaml")
	require.NoError(t, os.WriteFile(notADir, []byte("data"), 0644))

	var stderr bytes.Buffer
	code := Run([]string{planPath}, notADir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitPlanError, code)
	assert.Contains(t, stderr.String(), "not a directory")
}

func TestRun_RiskFilterMakesFeatureOrphan(t *testing.T) {
	tmpDir := t.TempDir()
	// Two plans in the YAML: one edge (excluded by --risk=beta?)
	// Wait, --risk=beta INCLUDES edge. We need --risk=candidate to
	// exclude the stable plan; then the stable feature file becomes
	// an orphan while the edge one is expected but missing.
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Edge Feature"
type: "functional"
status: "implemented"
risk: "edge"
scenarios:
  - "scenario"
---
feature: "Stable Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "stable_feature.feature", "Feature: Stable")

	var stderr bytes.Buffer
	// --risk=candidate excludes the stable plan, so the feature file
	// is now an orphan. (--risk=candidate INCLUDES edge, but we have
	// no feature file for it, so we'd also see a "missing"
	// classification.)
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{Risk: "candidate"}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "matches no test")
	assert.Contains(t, stderr.String(), `"stable_feature.feature"`)
}

func TestRun_StatusFilterMakesFeatureOrphan(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Planned Feature"
type: "functional"
status: "planned"
risk: "stable"
scenarios:
  - "scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "planned_feature.feature", "Feature: Planned")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{Status: "implemented"}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "matches no test")
}

func TestRun_TagFilterMakesFeatureOrphan(t *testing.T) {
	// With a tag filter active, a feature file whose plan does not
	// carry the requested tag is reported as an orphan, mirroring the
	// risk and status filter behavior.
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Tagged Feature"
type: "functional"
status: "planned"
risk: "stable"
tags:
  - single-node
scenarios:
  - "scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "tagged_feature.feature", "Feature: Tagged")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{Tags: []string{"multi-node"}}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitDiff, code)
	assert.Contains(t, stderr.String(), "matches no test")
}

func TestRun_ShowEmitsUnifiedDiffToStdout(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "login_feature.feature", "OLD LINE\n")

	var stdout, stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, true, &stdout, &stderr)
	assert.Equal(t, ExitDiff, code)
	out := stdout.String()
	assert.Contains(t, out, "--- a/login_feature.feature")
	assert.Contains(t, out, "+++ b/login_feature.feature")
	assert.Contains(t, out, "@@")
}

func TestRun_ShowWithNoDiffsProducesEmptyStdout(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	expected := common.GenerateGherkin(common.TestPlan{
		Feature: "Login Feature",
		Type:    "functional",
		Status:  "implemented",
		Risk:    "stable",
		Scenarios: []string{
			"User logs in\nGiven a user exists",
		},
	})
	writeFeature(t, featuresDir, "login_feature.feature", expected)

	var stdout, stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, true, &stdout, &stderr)
	assert.Equal(t, ExitClean, code)
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestRun_IgnoresNonFeatureFilesInFeaturesDir(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	expected := common.GenerateGherkin(common.TestPlan{
		Feature: "Login Feature",
		Type:    "functional",
		Status:  "implemented",
		Risk:    "stable",
		Scenarios: []string{
			"scenario",
		},
	})
	writeFeature(t, featuresDir, "login_feature.feature", expected)
	writeFeature(t, featuresDir, "notes.md", "# Notes")
	writeFeature(t, featuresDir, "README.txt", "hello")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitClean, code)
	assert.Empty(t, stderr.String())
}

func TestRun_IgnoresSubdirectoriesInFeaturesDir(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	expected := common.GenerateGherkin(common.TestPlan{
		Feature: "Login Feature",
		Type:    "functional",
		Status:  "implemented",
		Risk:    "stable",
		Scenarios: []string{
			"scenario",
		},
	})
	writeFeature(t, featuresDir, "login_feature.feature", expected)
	subDir := filepath.Join(featuresDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))
	writeFeature(t, subDir, "in_subdir.feature", "Feature: In Subdir")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &bytes.Buffer{}, &stderr)
	assert.Equal(t, ExitClean, code)
	assert.Empty(t, stderr.String())
}

func TestQuotedList(t *testing.T) {
	assert.Equal(t, "", quotedList(nil))
	assert.Equal(t, `"a"`, quotedList([]string{"a"}))
	assert.Equal(t, `"a", "b"`, quotedList([]string{"a", "b"}))
	assert.Equal(t, `"a", "b", "c"`, quotedList([]string{"a", "b", "c"}))
}

func TestReportMissing_EmptyProducesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	reportMissing(nil, &buf)
	assert.Empty(t, buf.String())
	buf.Reset()
	reportMissing([]classification{}, &buf)
	assert.Empty(t, buf.String())
}

func TestReportOrphans_EmptyProducesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	reportOrphans(nil, &buf)
	assert.Empty(t, buf.String())
	buf.Reset()
	reportOrphans([]string{}, &buf)
	assert.Empty(t, buf.String())
}

func TestReportDiffering_EmptyProducesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	reportDiffering(nil, &buf)
	assert.Empty(t, buf.String())
	buf.Reset()
	reportDiffering([]classification{}, &buf)
	assert.Empty(t, buf.String())
}

func TestIndexRendered_DetectsDuplicates(t *testing.T) {
	rendered := []common.RenderedFile{
		{Filename: "a.feature", Content: "x"},
		{Filename: "a.feature", Content: "y"},
	}
	indexed, dup := indexRendered(rendered)
	assert.Nil(t, indexed)
	assert.Equal(t, "a.feature", dup)
}

func TestIndexRendered_BuildsMapForUniqueNames(t *testing.T) {
	rendered := []common.RenderedFile{
		{Filename: "a.feature", Content: "x"},
		{Filename: "b.feature", Content: "y"},
	}
	indexed, dup := indexRendered(rendered)
	assert.Empty(t, dup)
	assert.Len(t, indexed, 2)
	assert.Equal(t, "x", indexed["a.feature"].Content)
	assert.Equal(t, "y", indexed["b.feature"].Content)
}

func TestScanFeatureDir_NonExistentDirReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	files, err := scanFeatureDir(filepath.Join(tmpDir, "missing"))
	assert.Error(t, err)
	assert.Nil(t, files)
}

func TestScanFeatureDir_PathIsFileReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "file.yaml")
	require.NoError(t, os.WriteFile(p, []byte("data"), 0644))
	files, err := scanFeatureDir(p)
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestReportMissing_PluralizesVerb(t *testing.T) {
	var buf bytes.Buffer
	reportMissing([]classification{
		{Filename: "a.feature"},
		{Filename: "b.feature"},
	}, &buf)
	assert.Contains(t, buf.String(), "are missing")

	buf.Reset()
	reportMissing([]classification{
		{Filename: "a.feature"},
	}, &buf)
	assert.Contains(t, buf.String(), "is missing")
	assert.NotContains(t, buf.String(), "are")
}

func TestReportOrphans_PluralizesVerb(t *testing.T) {
	var buf bytes.Buffer
	reportOrphans([]string{"a.feature", "b.feature"}, &buf)
	assert.Contains(t, buf.String(), "match no test")

	buf.Reset()
	reportOrphans([]string{"a.feature"}, &buf)
	assert.Contains(t, buf.String(), "matches no test")
}

func TestReportDiffering_PluralizesVerb(t *testing.T) {
	var buf bytes.Buffer
	reportDiffering([]classification{
		{Filename: "a.feature"},
		{Filename: "b.feature"},
	}, &buf)
	assert.Contains(t, buf.String(), "differ from")

	buf.Reset()
	reportDiffering([]classification{
		{Filename: "a.feature"},
	}, &buf)
	assert.Contains(t, buf.String(), "differs from")
}

func TestWriteUnifiedDiff_EmitsHeadersAndHunks(t *testing.T) {
	var buf bytes.Buffer
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)
	_ = planPath

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "login_feature.feature", "OLD\n")

	c := classification{
		Filename: "login_feature.feature",
		Content:  "NEW\n",
	}
	require.NoError(t, writeUnifiedDiff(&buf, featuresDir, c))
	out := buf.String()
	assert.Contains(t, out, "--- a/login_feature.feature")
	assert.Contains(t, out, "+++ b/login_feature.feature")
}

// alwaysFailWriter returns an error on every Write call.
type alwaysFailWriter struct{}

func (alwaysFailWriter) Write([]byte) (int, error) {
	return 0, assert.AnError
}

// failAfterNWriter succeeds for the first n Write calls and fails
// thereafter.  Counts both Fprintf and io.WriteString calls.
type failAfterNWriter struct {
	mu  sync.Mutex
	n   int
	got int
}

func (w *failAfterNWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.got++
	if w.got > w.n {
		return 0, assert.AnError
	}
	return len(p), nil
}

func TestWriteUnifiedDiff_FailOnSecondWrite(t *testing.T) {
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "x.feature", "OLD\n")

	w := &failAfterNWriter{n: 1}
	c := classification{Filename: "x.feature", Content: "NEW\n"}
	assert.Error(t, writeUnifiedDiff(w, featuresDir, c))
}

func TestWriteUnifiedDiff_FailOnThirdWrite(t *testing.T) {
	// Fail on the final io.WriteString call (after both Fprintf
	// headers).
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "x.feature", "OLD\n")

	w := &failAfterNWriter{n: 2}
	c := classification{Filename: "x.feature", Content: "NEW\n"}
	assert.Error(t, writeUnifiedDiff(w, featuresDir, c))
}

func TestWriteUnifiedDiff_FeatureFileVanishes(t *testing.T) {
	// Cover the os.ReadFile error branch by deleting the file
	// between scanFeatureDir and writeUnifiedDiff. We invoke the
	// function directly to control the timing.
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	path := filepath.Join(featuresDir, "x.feature")
	require.NoError(t, os.WriteFile(path, []byte("OLD"), 0644))
	require.NoError(t, os.Remove(path))

	c := classification{Filename: "x.feature", Content: "NEW"}
	assert.Error(t, writeUnifiedDiff(&bytes.Buffer{}, featuresDir, c))
}

func TestScanFeatureDir_ReadFileFailsForPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-denied test cannot run as root")
	}
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	path := filepath.Join(featuresDir, "x.feature")
	require.NoError(t, os.WriteFile(path, []byte("OLD"), 0000))
	t.Cleanup(func() { _ = os.Chmod(path, 0644) })

	files, err := scanFeatureDir(featuresDir)
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to read feature file")
}

func TestScanFeatureDir_ReadDirFailsForPermissionDenied(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("permission-denied test cannot run as root")
	}
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	require.NoError(t, os.Chmod(featuresDir, 0000))
	t.Cleanup(func() { _ = os.Chmod(featuresDir, 0755) })

	files, err := scanFeatureDir(featuresDir)
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to read features directory")
}

func TestWriteUnifiedDiff_PropagatesWriterErrors(t *testing.T) {
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "x.feature", "OLD\n")

	c := classification{Filename: "x.feature", Content: "NEW\n"}
	assert.Error(t, writeUnifiedDiff(alwaysFailWriter{}, featuresDir, c))
}

func TestRun_ShowDiffPropagatesWriterError(t *testing.T) {
	tmpDir := t.TempDir()
	planPath := writeYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - |
    User logs in
    Given a user exists
`)

	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))
	writeFeature(t, featuresDir, "login_feature.feature", "OLD\n")

	var stderr bytes.Buffer
	code := Run([]string{planPath}, featuresDir, common.FilterOptions{}, true, alwaysFailWriter{}, &stderr)
	assert.Equal(t, ExitPlanError, code)
	assert.Contains(t, stderr.String(), "failed to render diff")
}

func TestScanFeatureDir_HandlesUnreadableFile(t *testing.T) {
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	// Create a directory whose name ends in .feature so
	// os.ReadDir returns it as a non-FileInfo entry we cannot
	// read.
	subdir := filepath.Join(featuresDir, "fake.feature")
	require.NoError(t, os.MkdirAll(subdir, 0755))

	files, err := scanFeatureDir(featuresDir)
	// os.ReadDir calls os.File.ReadDir which descends into
	// directories only if they pass IsDir filtering. Our code
	// only reads files, so the subdirectory is silently skipped.
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestFindMissing_HandlesEmptyExpected(t *testing.T) {
	got := findMissing(map[string]classification{}, map[string]string{"x": "y"})
	assert.Empty(t, got)
}

func TestFindDiffering_HandlesEmptyExpected(t *testing.T) {
	got := findDiffering(map[string]classification{}, map[string]string{"x": "y"})
	assert.Empty(t, got)
}

func TestFindMissing_HandlesIdenticalNames(t *testing.T) {
	// Same filename in expected and actual → not missing.
	expected := map[string]classification{"x": {Filename: "x"}}
	actual := map[string]string{"x": "content"}
	got := findMissing(expected, actual)
	assert.Empty(t, got)
}

func TestFindDiffering_HandlesMatchingContent(t *testing.T) {
	expected := map[string]classification{"x": {Filename: "x", Content: "content"}}
	actual := map[string]string{"x": "content"}
	got := findDiffering(expected, actual)
	assert.Empty(t, got)
}
