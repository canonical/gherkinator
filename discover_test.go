package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverYAMLFiles_EmptyInputScansCwd(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.yaml"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "b.yml"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "c.txt"), []byte(""), 0644))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	files, err := DiscoverYAMLFiles(nil)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a.yaml", "b.yml"}, files)
}

func TestDiscoverYAMLFiles_EmptyArgsScansCwd(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "plan.yaml"), []byte(""), 0644))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	files, err := DiscoverYAMLFiles([]string{})
	require.NoError(t, err)
	assert.Equal(t, []string{"plan.yaml"}, files)
}

func TestDiscoverYAMLFiles_NoYAMLFilesInCwd(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte(""), 0644))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	require.NoError(t, os.Chdir(tmpDir))

	_, err = DiscoverYAMLFiles(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no YAML files found")
}

func TestDiscoverYAMLFiles_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yaml")
	require.NoError(t, os.WriteFile(inputFile, []byte(""), 0644))

	files, err := DiscoverYAMLFiles([]string{inputFile})
	require.NoError(t, err)
	assert.Equal(t, []string{inputFile}, files)
}

func TestDiscoverYAMLFiles_SingleYMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	inputFile := filepath.Join(tmpDir, "test-plan.yml")
	require.NoError(t, os.WriteFile(inputFile, []byte(""), 0644))

	files, err := DiscoverYAMLFiles([]string{inputFile})
	require.NoError(t, err)
	assert.Equal(t, []string{inputFile}, files)
}

func TestDiscoverYAMLFiles_SingleDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.yaml"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "b.yml"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "c.txt"), []byte(""), 0644))

	files, err := DiscoverYAMLFiles([]string{tmpDir})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		filepath.Join(tmpDir, "a.yaml"),
		filepath.Join(tmpDir, "b.yml"),
	}, files)
}

func TestDiscoverYAMLFiles_DirectoryWithNoYAML(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte(""), 0644))

	_, err := DiscoverYAMLFiles([]string{tmpDir})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no YAML files found in directory")
}

func TestDiscoverYAMLFiles_MixedFilesAndDirectories(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir1, "a.yaml"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "x.txt"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "b.yml"), []byte(""), 0644))

	explicitFile := filepath.Join(t.TempDir(), "c.yaml")
	require.NoError(t, os.WriteFile(explicitFile, []byte(""), 0644))

	files, err := DiscoverYAMLFiles([]string{dir1, explicitFile, dir2})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{
		filepath.Join(dir1, "a.yaml"),
		explicitFile,
		filepath.Join(dir2, "b.yml"),
	}, files)
}

func TestDiscoverYAMLFiles_NonexistentPath(t *testing.T) {
	_, err := DiscoverYAMLFiles([]string{"/nonexistent/path/to/file.yaml"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestDiscoverYAMLFiles_NonYAMLFile(t *testing.T) {
	tmpDir := t.TempDir()
	notYaml := filepath.Join(tmpDir, "notes.txt")
	require.NoError(t, os.WriteFile(notYaml, []byte(""), 0644))

	_, err := DiscoverYAMLFiles([]string{notYaml})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a YAML file")
}

func TestDiscoverYAMLFiles_Deduplicates(t *testing.T) {
	tmpDir := t.TempDir()
	yamlFile := filepath.Join(tmpDir, "plan.yaml")
	require.NoError(t, os.WriteFile(yamlFile, []byte(""), 0644))

	files, err := DiscoverYAMLFiles([]string{yamlFile, yamlFile})
	require.NoError(t, err)
	assert.Equal(t, []string{yamlFile}, files)
}

func TestDiscoverYAMLFiles_UnreadableDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; cannot drop read permission")
	}
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "locked"), 0000))
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(tmpDir, "locked"), 0755) })

	_, err := DiscoverYAMLFiles([]string{filepath.Join(tmpDir, "locked")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read directory")
}

func TestDiscoverYAMLFiles_SortsResults(t *testing.T) {
	dir := t.TempDir()
	c := filepath.Join(dir, "c.yaml")
	a := filepath.Join(dir, "a.yaml")
	b := filepath.Join(dir, "b.yaml")
	require.NoError(t, os.WriteFile(c, []byte(""), 0644))
	require.NoError(t, os.WriteFile(a, []byte(""), 0644))
	require.NoError(t, os.WriteFile(b, []byte(""), 0644))

	files, err := DiscoverYAMLFiles([]string{c, a, b})
	require.NoError(t, err)
	assert.Equal(t, []string{a, b, c}, files)
}

func TestDiscoverYAMLFiles_IgnoresSubdirectories(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.yaml"), []byte(""), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir", "nested.yaml"), []byte(""), 0644))

	files, err := DiscoverYAMLFiles([]string{tmpDir})
	require.NoError(t, err)
	assert.Equal(t, []string{filepath.Join(tmpDir, "a.yaml")}, files)
}
