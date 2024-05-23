package main

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	ConfigFileName = "osdctl"
	path           = "$HOME/.config"
)

type Config struct {
	pd_user_token string
	aws_proxy     string
}

func main() {

	LoadConfig(path)

}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("osdctl")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(viper.ConfigFileUsed())
	return
}
