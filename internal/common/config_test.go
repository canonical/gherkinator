package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitConfig_Defaults(t *testing.T) {
	viper.Reset()
	InitConfig()

	assert.Equal(t, "git", viper.GetString("tools.git"))
	assert.Equal(t, "python3", viper.GetString("tools.python3"))
	assert.Equal(t, "pip", viper.GetString("tools.pip"))
	assert.Equal(t, "make", viper.GetString("tools.make"))
	assert.Empty(t, viper.GetString("serve.dir"))
}

func TestInitConfig_EnvOverride(t *testing.T) {
	viper.Reset()
	t.Setenv("GHERKINATOR_TOOLS_GIT", "/usr/local/bin/git")
	InitConfig()

	assert.Equal(t, "/usr/local/bin/git", viper.GetString("tools.git"))
}

func TestInitConfig_NonHiddenConfigFallback(t *testing.T) {
	viper.Reset()
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// Strictly confined snaps cannot read hidden files, so a non-hidden
	// gherkinator.yaml must be honoured when no dotfile exists.
	config := "tools:\n  git: from-non-hidden\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "gherkinator.yaml"), []byte(config), 0o644))
	InitConfig()

	assert.Equal(t, "from-non-hidden", viper.GetString("tools.git"))
}

func TestInitConfig_HiddenConfigPrecedence(t *testing.T) {
	viper.Reset()
	tmpDir := t.TempDir()
	t.Chdir(tmpDir)

	// The hidden dotfile always wins over the non-hidden fallback so that
	// existing configurations keep their meaning.
	hidden := "tools:\n  git: from-hidden\n"
	visible := "tools:\n  git: from-visible\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".gherkinator.yaml"), []byte(hidden), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "gherkinator.yaml"), []byte(visible), 0o644))
	InitConfig()

	assert.Equal(t, "from-hidden", viper.GetString("tools.git"))
}
