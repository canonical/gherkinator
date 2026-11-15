package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetConfig reinitialises the global Viper instance so each test starts
// from production defaults (including AutomaticEnv) without config-file or
// environment leakage from other tests.
func resetConfig(t *testing.T) {
	t.Helper()
	viper.Reset()
	InitConfig()
	// The ambient environment may carry snap variables (e.g. when the IDE
	// itself is a snap); neutralise them so tests exercise the default path
	// unless they explicitly opt in below.
	t.Setenv("SNAP", "")
	t.Setenv("SNAP_USER_COMMON", "")
	t.Setenv("GHERKINATOR_SERVE_DIR", "")
}

func TestStagingDirFor_DefaultsToProjectDir(t *testing.T) {
	resetConfig(t)

	// No SNAP and no override: the classic in-project default is preserved
	// byte-for-byte so source installs behave exactly as before.
	dir, err := StagingDirFor("/home/user/project")
	require.NoError(t, err)
	assert.Equal(t, "/home/user/project/.gherkindocs", dir)
}

func TestStagingDirFor_SnapUsesUserCommon(t *testing.T) {
	resetConfig(t)
	t.Setenv("SNAP", "/snap/gherkinator/x1")
	t.Setenv("SNAP_USER_COMMON", "/home/user/snap/gherkinator/common")

	// Inside a snap the staging dir must move out of the project tree,
	// because the home interface denies access to hidden paths.
	dir, err := StagingDirFor("/home/user/project")
	require.NoError(t, err)
	assert.Equal(t, "/home/user/snap/gherkinator/common/gherkindocs", dir)
}

func TestStagingDirFor_SnapWithoutUserCommonFallsBack(t *testing.T) {
	resetConfig(t)
	t.Setenv("SNAP", "/snap/gherkinator/x1")
	t.Setenv("SNAP_USER_COMMON", "")

	// A missing SNAP_USER_COMMON is not expected inside a real snap, but
	// silently degrading to the project default is safer than panicking.
	dir, err := StagingDirFor("/home/user/project")
	require.NoError(t, err)
	assert.Equal(t, "/home/user/project/.gherkindocs", dir)
}

func TestStagingDirFor_EnvOverrideBeatsSnap(t *testing.T) {
	resetConfig(t)
	t.Setenv("SNAP", "/snap/gherkinator/x1")
	t.Setenv("SNAP_USER_COMMON", "/home/user/snap/gherkinator/common")
	t.Setenv("GHERKINATOR_SERVE_DIR", "/tmp/my-docs")

	// An explicit GHERKINATOR_SERVE_DIR wins even inside a snap.
	dir, err := StagingDirFor("/home/user/project")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/my-docs", dir)
}

func TestStagingDirFor_EnvOverrideResolvesRelativePath(t *testing.T) {
	resetConfig(t)
	t.Setenv("GHERKINATOR_SERVE_DIR", "reldocs")

	// Relative values resolve against the current working directory so the
	// resulting path is stable regardless of where the command runs from.
	dir, err := StagingDirFor("/home/user/project")
	require.NoError(t, err)
	abs, err := filepath.Abs("reldocs")
	require.NoError(t, err)
	assert.Equal(t, abs, dir)
}

func TestStagingDirFor_WhitespaceOverrideIgnored(t *testing.T) {
	resetConfig(t)
	t.Setenv("GHERKINATOR_SERVE_DIR", "   ")

	// Whitespace-only values are treated as unset rather than as a broken
	// relative path.
	dir, err := StagingDirFor("/home/user/project")
	require.NoError(t, err)
	assert.Equal(t, "/home/user/project/.gherkindocs", dir)
}

func TestStagingDirFor_ViperSetOverride(t *testing.T) {
	resetConfig(t)
	// viper.Set is the highest-priority source; it must beat the snap
	// default just like the environment variable does.
	viper.Set("serve.dir", "/opt/docs-staging")
	t.Setenv("SNAP", "/snap/gherkinator/x1")
	t.Setenv("SNAP_USER_COMMON", "/home/user/snap/gherkinator/common")

	dir, err := StagingDirFor("/home/user/project")
	require.NoError(t, err)
	assert.Equal(t, "/opt/docs-staging", dir)
}

func TestStagingDirFor_ResolutionFailure(t *testing.T) {
	resetConfig(t)
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)
	require.NoError(t, os.Remove(tmpDir))
	t.Setenv("GHERKINATOR_SERVE_DIR", "relative-docs")

	// A relative override resolves via the working directory; with the
	// cwd deleted, resolution fails instead of returning a bogus path.
	_, err := StagingDirFor("/anywhere")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve serve directory")
}
