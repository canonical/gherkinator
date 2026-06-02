package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DiscoverYAMLFiles resolves a list of user-supplied paths into a deduplicated,
// sorted list of .yaml or .yml file paths.
//
// If paths is empty, the current directory (".") is scanned.  Each entry in
// paths may be either a YAML file (included as-is) or a directory (scanned
// non-recursively for YAML files).  An error is returned if a path does not
// exist or if a directory contains no YAML files.
func DiscoverYAMLFiles(paths []string) ([]string, error) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	seen := make(map[string]bool)
	var files []string
	var errs []error

	add := func(p string) {
		if seen[p] {
			return
		}
		seen[p] = true
		files = append(files, p)
	}

	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", p, err))
			continue
		}

		if info.IsDir() {
			entries, err := os.ReadDir(p)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to read directory %s: %w", p, err))
				continue
			}
			found := false
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				if HasYAMLExt(e.Name()) {
					add(filepath.Join(p, e.Name()))
					found = true
				}
			}
			if !found {
				errs = append(errs, fmt.Errorf("no YAML files found in directory %s", p))
			}
			continue
		}

		if !HasYAMLExt(p) {
			errs = append(errs, fmt.Errorf("%s: not a YAML file", p))
			continue
		}

		add(p)
	}

	if errs != nil {
		return nil, errors.Join(errs...)
	}

	sort.Strings(files)
	return files, nil
}

// HasYAMLExt reports whether name has a .yaml or .yml file extension
// (case-insensitive).
func HasYAMLExt(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml")
}
