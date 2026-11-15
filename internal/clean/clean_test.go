package clean

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

func TestCleanDirectory_RemovesFeatureFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .feature files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("Feature: Login"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "signup.feature"), []byte("Feature: Signup"), 0644))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "signup.feature"))
}

func TestCleanDirectory_RemovesMarkdownFiles(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.md"), []byte("# Login"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "signup.md"), []byte("# Signup"), 0644))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(tmpDir, "login.md"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "signup.md"))
}

func TestCleanDirectory_RemovesServeDirectory(t *testing.T) {
	resetStagingConfig(t)
	tmpDir := t.TempDir()

	serveDir := filepath.Join(tmpDir, ".gherkindocs")
	require.NoError(t, os.MkdirAll(serveDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(serveDir, "index.html"), []byte("<html>"), 0644))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoDirExists(t, serveDir)
}

func TestCleanDirectory_PreservesYAMLFiles(t *testing.T) {
	tmpDir := t.TempDir()

	yamlFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(yamlFile, []byte("feature: test"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("Feature: Login"), 0644))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	// YAML file should still exist
	assert.FileExists(t, yamlFile)
	// Feature file should be removed
	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
}

func TestCleanDirectory_PreservesGoFiles(t *testing.T) {
	tmpDir := t.TempDir()

	goFile := filepath.Join(tmpDir, "main.go")
	require.NoError(t, os.WriteFile(goFile, []byte("package main"), 0644))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.FileExists(t, goFile)
}

func TestCleanDirectory_SkipsSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.MkdirAll(subDir, 0755))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.DirExists(t, subDir)
}

func TestCleanDirectory_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)
}

func TestCleanDirectory_MixedFiles(t *testing.T) {
	resetStagingConfig(t)
	tmpDir := t.TempDir()

	// Create various files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("feature"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs.md"), []byte("# Docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test-plan.yaml"), []byte("yaml"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("go"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, ".gherkindocs"), 0755))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "docs.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "test-plan.yaml"))
	assert.FileExists(t, filepath.Join(tmpDir, "main.go"))
	assert.NoDirExists(t, filepath.Join(tmpDir, ".gherkindocs"))
}

func TestCleanDirectory_NonexistentDirectory(t *testing.T) {
	err := CleanDirectory("/nonexistent/directory")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read directory")
}

func TestCleanDirectory_GeneratedFileRemovalFails(t *testing.T) {
	resetStagingConfig(t)
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("feature"), 0o644))

	// A read-only target directory makes the generated-file removal fail;
	// that is a real error and must abort the command.
	require.NoError(t, os.Chmod(tmpDir, 0o555))
	t.Cleanup(func() {
		_ = os.Chmod(tmpDir, 0o755)
	})

	err := CleanDirectory(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove")
}

func TestCleanDirectory_CaseInsensitiveExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.FEATURE"), []byte("feature"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs.MD"), []byte("docs"), 0o644))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(tmpDir, "login.FEATURE"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "docs.MD"))
}

// resetStagingConfig reinitialises Viper so staging-dir overrides from other
// tests or the environment do not leak into these cases.
func resetStagingConfig(t *testing.T) {
	t.Helper()
	viper.Reset()
	common.InitConfig()
	// The ambient environment may carry snap variables (e.g. when the IDE
	// itself is a snap); neutralise them so tests exercise the default path
	// unless they explicitly opt in below.
	t.Setenv("SNAP", "")
	t.Setenv("SNAP_USER_COMMON", "")
	t.Setenv("GHERKINATOR_SERVE_DIR", "")
}

func TestCleanDirectory_RemovesEnvStagingDir(t *testing.T) {
	resetStagingConfig(t)
	tmpDir := t.TempDir()
	customDir := filepath.Join(tmpDir, "custom-docs")
	require.NoError(t, os.MkdirAll(customDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("feature"), 0o644))

	// GHERKINATOR_SERVE_DIR redirects the staging dir outside the project;
	// clean must resolve the same path serve wrote to.
	t.Setenv("GHERKINATOR_SERVE_DIR", customDir)
	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoDirExists(t, customDir)
	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
}

func TestCleanDirectory_RemovesSnapStagingDir(t *testing.T) {
	resetStagingConfig(t)
	tmpDir := t.TempDir()
	snapCommon := filepath.Join(tmpDir, "snap-common")
	stagingDir := filepath.Join(snapCommon, "gherkindocs")
	require.NoError(t, os.MkdirAll(stagingDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("feature"), 0o644))

	// Inside a snap the staging dir is shared under $SNAP_USER_COMMON and
	// clean removes it regardless of the -d target.
	t.Setenv("SNAP", "/snap/gherkinator/x1")
	t.Setenv("SNAP_USER_COMMON", snapCommon)
	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoDirExists(t, stagingDir)
	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
}

func TestCleanDirectory_StagingResolutionFails(t *testing.T) {
	resetStagingConfig(t)
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	require.NoError(t, os.Remove(tmpDir))
	t.Setenv("GHERKINATOR_SERVE_DIR", "relative-docs")

	// A relative override resolves via the working directory; with the
	// cwd deleted, clean reports the resolution failure instead of
	// silently cleaning the wrong staging directory.
	err := CleanDirectory(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve serve directory")
}

func TestCleanDirectory_StagingRemovalWarnsAndContinues(t *testing.T) {
	resetStagingConfig(t)
	tmpDir := t.TempDir()
	stagingDir := filepath.Join(tmpDir, ".gherkindocs")
	lockedDir := filepath.Join(stagingDir, "locked")
	require.NoError(t, os.MkdirAll(lockedDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(lockedDir, "file.txt"), []byte("data"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("feature"), 0o644))

	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}
	// A read-only subdirectory makes RemoveAll fail partway through; clean
	// must warn instead of aborting so the generated files still get removed.
	require.NoError(t, os.Chmod(lockedDir, 0o555))
	t.Cleanup(func() {
		_ = os.Chmod(lockedDir, 0o755)
	})

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	cleanErr := CleanDirectory(tmpDir)

	require.NoError(t, w.Close())
	os.Stderr = oldStderr
	warnings, err := io.ReadAll(r)
	require.NoError(t, err)

	// The staging failure is a warning, not a command failure.
	assert.NoError(t, cleanErr)
	assert.Contains(t, string(warnings), "warning: could not fully remove serve directory")
	assert.Contains(t, string(warnings), stagingDir)

	// The generated files must still have been removed.
	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
}
