package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/gherkinator/internal/common"
	"github.com/canonical/gherkinator/internal/diff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeDiffYAML(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(p, []byte(body), 0644))
	return p
}

func writeDiffFeature(t *testing.T, dir, name, body string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0644))
}

func TestDiffCommand_HasCorrectUse(t *testing.T) {
	resetFlags()

	for _, c := range rootCmd.Commands() {
		if c.Name() == "diff" {
			assert.Contains(t, c.Use, "diff")
			assert.Contains(t, c.Short, "Diff")
			return
		}
	}
	t.Fatal("diff command not registered")
}

func TestDiffCommand_RequiresExactArgs(t *testing.T) {
	resetFlags()

	// No args.
	rootCmd.SetArgs([]string{"diff"})
	err := rootCmd.Execute()
	assert.Error(t, err)

	resetFlags()

	// One arg only.
	rootCmd.SetArgs([]string{"diff", "plans/"})
	err = rootCmd.Execute()
	assert.Error(t, err)

	resetFlags()

	// Three args.
	rootCmd.SetArgs([]string{"diff", "plans/", "features/", "extra"})
	err = rootCmd.Execute()
	assert.Error(t, err)
}

func TestDiffCommand_InvalidRiskFlagReturnsError(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	planPath := writeDiffYAML(t, tmpDir, "p.yaml", `feature: "F"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	rootCmd.SetArgs([]string{"diff", "--risk", "bogus", planPath, featuresDir})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--risk must be one of")
}

func TestDiffCommand_InvalidStatusFlagReturnsError(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	planPath := writeDiffYAML(t, tmpDir, "p.yaml", `feature: "F"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	rootCmd.SetArgs([]string{"diff", "--status", "bogus", planPath, featuresDir})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--status must be one of")
}

func TestDiffCommand_NonexistentPlansArgReturnsError(t *testing.T) {
	resetFlags()
	tmpDir := t.TempDir()
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	rootCmd.SetArgs([]string{"diff", filepath.Join(tmpDir, "nonexistent.yaml"), featuresDir})
	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestDiffCommand_FlagsRegistered(t *testing.T) {
	resetFlags()

	for _, c := range rootCmd.Commands() {
		if c.Name() != "diff" {
			continue
		}
		flags := c.Flags()
		assert.NotNil(t, flags.Lookup("risk"))
		assert.NotNil(t, flags.Lookup("status"))
		assert.NotNil(t, flags.Lookup("tag"))
		assert.NotNil(t, flags.Lookup("show"))
		return
	}
	t.Fatal("diff command not registered")
}

func TestDiffCommand_InvalidTagFlagReturnsError(t *testing.T) {
	// --tag validation happens before any disk I/O, so the tag value is
	// refused before the plans or features directory are read.
	resetFlags()
	tmpDir := t.TempDir()
	planPath := writeDiffYAML(t, tmpDir, "p.yaml", `feature: "Foo"
type: "functional"
status: "implemented"
risk: "stable"
scenarios:
  - "scenario"
`)
	featuresDir := filepath.Join(tmpDir, "features")
	require.NoError(t, os.MkdirAll(featuresDir, 0755))

	rootCmd.SetArgs([]string{"diff", "--tag", "foo bar", planPath, featuresDir})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--tag")
	assert.Contains(t, err.Error(), "invalid tag 'foo bar'")
}

func TestDiffExit_InvokesRunWithProvidedArgs(t *testing.T) {
	// The cobra RunE for `diff` calls os.Exit directly; we cannot
	// exercise it here without aborting the test. This test confirms
	// the diff.Run entrypoint returns 0 for a clean diff (so we
	// know the cobra command would exit 0 too); full behavioural
	// coverage lives in internal/diff/diff_test.go.
	tmpDir := t.TempDir()
	planPath := writeDiffYAML(t, tmpDir, "p.yaml", `feature: "Login Feature"
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
	content := common.GenerateGherkin(common.TestPlan{
		Feature: "Login Feature",
		Type:    "functional",
		Status:  "implemented",
		Risk:    "stable",
		Scenarios: []string{
			"User logs in\nGiven a user exists",
		},
	})
	writeDiffFeature(t, featuresDir, "login_feature.feature", content)

	var stdout, stderr bytes.Buffer
	code := diff.Run([]string{planPath}, featuresDir, common.FilterOptions{}, false, &stdout, &stderr)
	assert.Equal(t, 0, code)
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}
