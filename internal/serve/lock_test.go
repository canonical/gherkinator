package serve

import (
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockServeInstance_UsesSnapCommon(t *testing.T) {
	// The default lock location inside a snap is $SNAP_USER_COMMON.
	snapCommon := t.TempDir()
	t.Setenv("SNAP", "/snap/gherkinator/x1")
	t.Setenv("SNAP_USER_COMMON", snapCommon)

	release, err := lockServeInstance()
	require.NoError(t, err)
	assert.FileExists(t, filepath.Join(snapCommon, "serve.lock"))
	release()
}

func TestLockServeFile_Exclusive(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "serve.lock")

	release, err := lockServeFile(lockPath)
	require.NoError(t, err)

	// A second instance must be refused while the lock is held.
	_, err = lockServeFile(lockPath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// After release the lock is available again.
	release()
	_, err = lockServeFile(lockPath)
	require.NoError(t, err)
}

func TestLockServeFile_ParentIsFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "regular-file")
	require.NoError(t, os.WriteFile(file, []byte("data"), 0o644))

	_, err := lockServeFile(filepath.Join(file, "serve.lock"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create serve lock directory")
}

func TestLockServeFile_UnwritableDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}

	dir := t.TempDir()
	require.NoError(t, os.Chmod(dir, 0o555))
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o755)
	})

	_, err := lockServeFile(filepath.Join(dir, "serve.lock"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open serve lock")
}

func TestLockServeFile_ReleasedOnClose(t *testing.T) {
	// The kernel drops the flock when the file is closed or the process
	// dies; simulate the death path by closing without the release func.
	lockPath := filepath.Join(t.TempDir(), "serve.lock")

	release, err := lockServeFile(lockPath)
	require.NoError(t, err)
	release()

	_, err = lockServeFile(lockPath)
	require.NoError(t, err)
}

func TestServeLockPath_Snap(t *testing.T) {
	t.Setenv("SNAP", "/snap/gherkinator/x1")
	t.Setenv("SNAP_USER_COMMON", "/home/user/snap/gherkinator/common")

	path, err := serveLockPath()
	require.NoError(t, err)
	assert.Equal(t, "/home/user/snap/gherkinator/common/serve.lock", path)
}

func TestServeLockPath_UserCache(t *testing.T) {
	t.Setenv("SNAP", "")
	t.Setenv("SNAP_USER_COMMON", "")
	cache := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cache)

	path, err := serveLockPath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(cache, "gherkinator", "serve.lock"), path)
}

func TestServeLockPath_NoCacheDir(t *testing.T) {
	t.Setenv("SNAP", "")
	t.Setenv("SNAP_USER_COMMON", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "")

	// With no resolvable cache directory the lock path cannot be
	// determined and serve must fail with a clear error instead of
	// silently skipping the mutual exclusion.
	_, err := serveLockPath()
	require.Error(t, err)
}

func TestLockServeInstance_UnresolvableLockPath(t *testing.T) {
	t.Setenv("SNAP", "")
	t.Setenv("SNAP_USER_COMMON", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "")

	_, err := lockServeInstance()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve serve lock path")
}

func TestCheckServingPortFree_ListenerPresent(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = listener.Close()
	})

	err = checkServingPortFree("http://" + listener.Addr().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already in use")
}

func TestCheckServingPortFree_Free(t *testing.T) {
	// Reserve a port and release it so nothing is listening; the probe
	// must report the port as free.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())

	assert.NoError(t, checkServingPortFree("http://"+addr))
}
