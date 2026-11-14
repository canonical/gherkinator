package serve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnsureWritableDir_CreatesMissingDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")

	// The serve directory is frequently configured to a path that does not
	// exist yet; the probe must create it rather than fail.
	require.NoError(t, ensureWritableDir(dir))
	assert.DirExists(t, dir)

	// The probe file must not linger after the check.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestEnsureWritableDir_ExistingDirectory(t *testing.T) {
	dir := t.TempDir()

	require.NoError(t, ensureWritableDir(dir))
}

func TestEnsureWritableDir_PathIsFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "regular-file")
	require.NoError(t, os.WriteFile(file, []byte("data"), 0o644))

	// A file where a directory is expected is a configuration mistake and
	// must produce an error mentioning the serve directory.
	err := ensureWritableDir(file)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "serve directory")
}

func TestEnsureWritableDir_ReadOnlyDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}

	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o755)
	})

	// A read-only serve directory (e.g. inside the read-only snap tree)
	// must fail the probe so the caller can suggest GHERKINATOR_SERVE_DIR.
	err := ensureWritableDir(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not writable")
}
