package serve

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/canonical/gherkinator/internal/common"
)

// resetServeConfig reinitialises Viper with production defaults and
// neutralises ambient snap variables so each test exercises the staging
// path it explicitly opts into.
func resetServeConfig(t *testing.T) {
	t.Helper()
	viper.Reset()
	common.InitConfig()
	t.Setenv("SNAP", "")
	t.Setenv("SNAP_USER_COMMON", "")
	t.Setenv("GHERKINATOR_SERVE_DIR", "")
}

// writeFakeGit stages an executable that stands in for git so prepareServeSite
// can be tested without network access.
func writeFakeGit(t *testing.T, script string) string {
	t.Helper()

	binDir := t.TempDir()
	gitPath := filepath.Join(binDir, "fake-git")
	require.NoError(t, os.WriteFile(gitPath, []byte(script), 0o755))
	viper.Set("tools.git", gitPath)
	return gitPath
}

// writePlanInput creates a minimal valid test plan file.
func writePlanInput(t *testing.T, dir string) string {
	t.Helper()

	plan := "feature: \"Login\"\ntype: functional\nstatus: planned\nrisk: stable\nscenarios:\n  - \"Login works\"\n"
	path := filepath.Join(dir, "plan.yaml")
	require.NoError(t, os.WriteFile(path, []byte(plan), 0o644))
	return path
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

func TestPrepareServeSite_StagingClearWarns(t *testing.T) {
	resetServeConfig(t)
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}
	cwd := t.TempDir()
	input := writePlanInput(t, cwd)

	// Locked leftover content inside the staging dir makes the initial
	// RemoveAll fail partway; the run must warn and still complete.
	stagingDir := filepath.Join(cwd, ".gherkindocs")
	lockedDir := filepath.Join(stagingDir, "locked")
	require.NoError(t, os.MkdirAll(lockedDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(lockedDir, "file.txt"), []byte("data"), 0o644))
	require.NoError(t, os.Chmod(lockedDir, 0o555))
	t.Cleanup(func() {
		_ = os.Chmod(lockedDir, 0o755)
	})

	writeFakeGit(t, "#!/bin/sh\ndest=\"$3\"\nmkdir -p \"$dest/docs\"\nprintf 'project = \"placeholder\"\\n' > \"$dest/docs/conf.py\"\n")

	var docsDir string
	var regenerateDocs func() error
	stderr := captureStderr(t, func() {
		var err error
		docsDir, regenerateDocs, err = prepareServeSite(cwd, []string{input}, "", common.FilterOptions{})
		require.NoError(t, err)
	})

	assert.Contains(t, stderr, "warning: could not fully clear serve directory")
	assert.Contains(t, stderr, stagingDir)
	assert.NotEmpty(t, docsDir)

	// Despite the partial wipe failure the site must still be usable.
	require.NoError(t, regenerateDocs())
}

func TestPrepareServeSite_PruneFails(t *testing.T) {
	resetServeConfig(t)
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}
	cwd := t.TempDir()
	input := writePlanInput(t, cwd)

	// A template dir containing a read-only subdirectory defeats the prune
	// step and must surface as a wrapped prune failure.
	writeFakeGit(t, "#!/bin/sh\ndest=\"$3\"\nmkdir -p \"$dest/docs/how-to/locked\"\ntouch \"$dest/docs/how-to/locked/file\"\nchmod 555 \"$dest/docs/how-to/locked\"\n")

	// Restore the mode so t.TempDir cleanup can remove the tree.
	lockedDir := filepath.Join(cwd, ".gherkindocs", "docs", "how-to", "locked")
	t.Cleanup(func() {
		_ = os.Chmod(lockedDir, 0o755)
	})

	_, _, err := prepareServeSite(cwd, []string{input}, "", common.FilterOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prune sphinx-stack defaults")
}

func TestPrepareServeSite_InvalidPlanFails(t *testing.T) {
	resetServeConfig(t)
	cwd := t.TempDir()

	writeFakeGit(t, "#!/bin/sh\ndest=\"$3\"\nmkdir -p \"$dest/docs\"\nprintf 'project = \"placeholder\"\\n' > \"$dest/docs/conf.py\"\n")

	// An invalid plan must abort the site preparation before the TUI.
	badPlan := filepath.Join(cwd, "bad.yaml")
	require.NoError(t, os.WriteFile(badPlan, []byte("{{{not yaml"), 0o644))

	_, _, err := prepareServeSite(cwd, []string{badPlan}, "", common.FilterOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare sphinx site")
	assert.Contains(t, err.Error(), "failed to generate sphinx docs for")
}

func TestPrepareServeSite_ConfPyMissingProjectFails(t *testing.T) {
	resetServeConfig(t)
	cwd := t.TempDir()
	input := writePlanInput(t, cwd)

	// A conf.py without a project line must abort the site preparation.
	writeFakeGit(t, "#!/bin/sh\ndest=\"$3\"\nmkdir -p \"$dest/docs\"\nprintf 'author = \"x\"\\n' > \"$dest/docs/conf.py\"\n")

	_, _, err := prepareServeSite(cwd, []string{input}, "", common.FilterOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare sphinx site")
	assert.Contains(t, err.Error(), "could not find 'project' variable")
}

func TestPrepareServeSite_EndToEnd(t *testing.T) {
	resetServeConfig(t)
	cwd := t.TempDir()

	// The fake git mimics a sphinx-stack clone closely enough for the
	// pipeline: a docs/ directory with a conf.py and a default index.rst.
	writeFakeGit(t, "#!/bin/sh\ndest=\"$3\"\nmkdir -p \"$dest/docs\"\nprintf 'project = \"placeholder\"\\n' > \"$dest/docs/conf.py\"\nprintf 'index\\n' > \"$dest/docs/index.rst\"\n")

	input := writePlanInput(t, cwd)
	docsDir, regenerateDocs, err := prepareServeSite(cwd, []string{input}, "My Project", common.FilterOptions{})
	require.NoError(t, err)

	// The docs live under the in-project staging default.
	assert.Equal(t, filepath.Join(cwd, ".gherkindocs", "docs"), docsDir)

	// The project name must have been rewritten into conf.py.
	conf, err := os.ReadFile(filepath.Join(docsDir, "conf.py"))
	require.NoError(t, err)
	assert.Contains(t, string(conf), "project = \"My Project\"")

	// The generated plan page must exist next to the type landing page.
	typeDir := filepath.Join(docsDir, "functional")
	assert.FileExists(t, filepath.Join(typeDir, "login.md"))
	assert.FileExists(t, filepath.Join(typeDir, "index.md"))

	// The generated plan must be regenerable.
	require.NoError(t, regenerateDocs())
}

func TestPrepareServeSite_CloneFails(t *testing.T) {
	resetServeConfig(t)
	cwd := t.TempDir()
	input := writePlanInput(t, cwd)

	// A failing clone must name the staging directory and the remediation.
	writeFakeGit(t, "#!/bin/sh\nexit 1\n")

	_, _, err := prepareServeSite(cwd, []string{input}, "", common.FilterOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to clone sphinx-stack into")
	assert.Contains(t, err.Error(), "GHERKINATOR_SERVE_DIR")
}

func TestPrepareServeSite_UnwritableStagingDir(t *testing.T) {
	resetServeConfig(t)
	cwd := t.TempDir()
	input := writePlanInput(t, cwd)

	// Pointing the serve directory at a regular file must fail the
	// pre-flight check with the configuration hint.
	fileStaging := filepath.Join(cwd, "not-a-dir")
	require.NoError(t, os.WriteFile(fileStaging, []byte("data"), 0o644))
	t.Setenv("GHERKINATOR_SERVE_DIR", fileStaging)

	_, _, err := prepareServeSite(cwd, []string{input}, "", common.FilterOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "serve directory")
	assert.Contains(t, err.Error(), "GHERKINATOR_SERVE_DIR")
}

func TestPrepareServeSite_StagingResolutionFails(t *testing.T) {
	resetServeConfig(t)
	cwd := t.TempDir()

	// A relative override resolves via the working directory; with the
	// cwd deleted, resolution fails with a dedicated error. No input
	// files are needed because resolution fails before regeneration.
	t.Chdir(cwd)
	require.NoError(t, os.Remove(cwd))
	t.Setenv("GHERKINATOR_SERVE_DIR", "relative-docs")

	_, _, err := prepareServeSite(cwd, nil, "", common.FilterOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve serve directory")
}
