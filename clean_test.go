package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCleanDirectory_CaseInsensitiveExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.FEATURE"), []byte("feature"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs.MD"), []byte("docs"), 0644))

	err := CleanDirectory(tmpDir)
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(tmpDir, "login.FEATURE"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "docs.MD"))
}

func TestCleanCommand_Success(t *testing.T) {
	setupCommands()
	tmpDir := t.TempDir()

	// Create files to clean
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "login.feature"), []byte("feature"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docs.md"), []byte("docs"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "test-plan.yaml"), []byte("yaml"), 0644))

	rootCmd.SetArgs([]string{"clean", "-d", tmpDir})
	err := rootCmd.Execute()
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(tmpDir, "login.feature"))
	assert.NoFileExists(t, filepath.Join(tmpDir, "docs.md"))
	assert.FileExists(t, filepath.Join(tmpDir, "test-plan.yaml"))
}

func TestCleanCommand_NonexistentDir(t *testing.T) {
	setupCommands()
	rootCmd.SetArgs([]string{"clean", "-d", "/nonexistent/dir"})
	err := rootCmd.Execute()
	assert.Error(t, err)
}
