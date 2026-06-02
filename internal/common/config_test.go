package common

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitConfig_Defaults(t *testing.T) {
	viper.Reset()
	InitConfig()

	assert.Equal(t, "git", viper.GetString("tools.git"))
	assert.Equal(t, "python3", viper.GetString("tools.python3"))
	assert.Equal(t, "pip", viper.GetString("tools.pip"))
	assert.Equal(t, "make", viper.GetString("tools.make"))
}

func TestInitConfig_EnvOverride(t *testing.T) {
	viper.Reset()
	t.Setenv("GHERKINATOR_TOOLS_GIT", "/usr/local/bin/git")
	InitConfig()

	assert.Equal(t, "/usr/local/bin/git", viper.GetString("tools.git"))
}
