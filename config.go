package main

import (
	"strings"

	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetDefault("tools.git", "git")
	viper.SetDefault("tools.python3", "python3")
	viper.SetDefault("tools.pip", "pip")
	viper.SetDefault("tools.make", "make")

	viper.SetConfigName(".gherkinator")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("gherkinator")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()
}
