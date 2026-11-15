// Package common provides shared types and utilities for the gherkinator CLI.
package common

import (
	"errors"
	"strings"

	"github.com/spf13/viper"
)

// InitConfig sets up Viper with default tool paths and binds it to the
// .gherkinator.yaml config file in the current or home directory. It also
// enables automatic environment variable overrides using the GHERKINATOR_
// prefix (with dots in keys replaced by underscores).
func InitConfig() {
	viper.SetDefault("tools.git", "git")
	viper.SetDefault("tools.python3", "python3")
	viper.SetDefault("tools.pip", "pip")
	viper.SetDefault("tools.make", "make")
	viper.SetDefault("serve.dir", "")

	viper.SetConfigName(".gherkinator")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	viper.SetEnvPrefix("gherkinator")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// The hidden dotfile always wins. Fall back to a non-hidden
	// gherkinator.yaml because strictly confined snaps cannot read hidden
	// files inside the project tree or the home directory.
	err := viper.ReadInConfig()
	if err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			viper.SetConfigName("gherkinator")
			_ = viper.ReadInConfig()
		}
	}
}
