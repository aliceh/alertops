package main

import (
	"fmt"

	pagerduty "github.com/aliceh/alertops/pkg/pagerduty"

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

	config, err := LoadConfig(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	pdclient := pagerduty.NewClient().WithOauthToken(config.pd_user_token)
	pdclient.GetPDServiceIDs()

}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("osdctl")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	config.pd_user_token = viper.GetString("pd_user_token")
	config.aws_proxy = viper.GetString("aws_proxy")

	return config, nil
}
