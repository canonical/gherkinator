package serve

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// serveLockFileName is the name of the lock file that serializes serve
// invocations for the current user.
const serveLockFileName = "serve.lock"

// serveLockPath returns the path of the serve lock file. It lives outside
// the staging directory because the staging directory is wiped on every
// run; a stable path keeps the lock's inode intact across invocations.
func serveLockPath() (string, error) {
	if snapCommon := os.Getenv("SNAP_USER_COMMON"); os.Getenv("SNAP") != "" && snapCommon != "" {
		return filepath.Join(snapCommon, serveLockFileName), nil
	}

	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cache, "gherkinator", serveLockFileName), nil
}

// lockServeInstance acquires the serve lock at its default location for
// the current user. The returned func releases it.
func lockServeInstance() (func(), error) {
	path, err := serveLockPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve serve lock path: %w", err)
	}
	return lockServeFile(path)
}

// lockServeFile acquires an exclusive, kernel-managed lock on path that
// serializes serve invocations. The lock is held by the returned file for
// the lifetime of the process and is released by the kernel on any exit,
// including SIGKILL, so it can never go stale.
func lockServeFile(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create serve lock directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open serve lock %s: %w", path, err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("another gherkinator serve instance is already running (serve lock %s is held); stop it first", path)
	}

	return func() { _ = f.Close() }, nil
}
