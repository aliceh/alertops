package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	ConfigFileName = "srepd"
	ConfigFileType = "yaml"
	Path           = "$HOME/.config/srepd"
)

type Config struct {
	Token        string
	Teams        []string
	SilentUser   string
	IgnoredUsers []string
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType(ConfigFileType)

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		return config, err
	}
	config.Token = viper.GetString("token")
	config.Teams = viper.GetStringSlice("teams")
	config.SilentUser = viper.GetString("silentuser")
	config.IgnoredUsers = viper.GetStringSlice("ignoredusers")

	return config, nil
}
