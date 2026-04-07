package main

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInitConfig_Defaults(t *testing.T) {
	viper.Reset()
	initConfig()

	assert.Equal(t, "git", viper.GetString("tools.git"))
	assert.Equal(t, "python3", viper.GetString("tools.python3"))
	assert.Equal(t, "pip", viper.GetString("tools.pip"))
	assert.Equal(t, "make", viper.GetString("tools.make"))
}

func TestInitConfig_EnvOverride(t *testing.T) {
	viper.Reset()
	t.Setenv("GHERKINATOR_TOOLS_GIT", "/usr/local/bin/git")
	initConfig()

	assert.Equal(t, "/usr/local/bin/git", viper.GetString("tools.git"))
}
