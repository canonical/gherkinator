package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// snapStagingDirName is the directory name used for the shared staging
// directory under $SNAP_USER_COMMON when running inside a snap.
const snapStagingDirName = "gherkindocs"

// StagingDirFor returns the directory used to stage generated documentation
// for the `serve` command and removed again by the `clean` command.
//
// Precedence (first match wins):
//  1. the "serve.dir" configuration key, which can also be set through the
//     GHERKINATOR_SERVE_DIR environment variable; relative values resolve
//     against the current working directory;
//  2. $SNAP_USER_COMMON/gherkindocs when running inside a snap, because a
//     strictly confined snap cannot write the hidden paths (the staging
//     directory, the virtualenv, the doctrees) inside the project tree;
//  3. <projectDir>/.gherkindocs, the classic default.
func StagingDirFor(projectDir string) (string, error) {
	if dir := strings.TrimSpace(viper.GetString("serve.dir")); dir != "" {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve serve directory %q: %w", dir, err)
		}
		return abs, nil
	}

	if os.Getenv("SNAP") != "" {
		if snapCommon := os.Getenv("SNAP_USER_COMMON"); snapCommon != "" {
			return filepath.Join(snapCommon, snapStagingDirName), nil
		}
	}

	return filepath.Join(projectDir, ".gherkindocs"), nil
}
