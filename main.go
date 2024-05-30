package main

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	pd "github.com/aliceh/alertops/pkg/pagerduty"

	"github.com/spf13/viper"
)

const (
	ConfigFileName = "osdctl"
	path           = "$HOME/.config"
)

type Config struct {
	pd_user_token string
	teams         []string
	silentUser    string
	ignoredUsers  []string
}

func main() {

	config, _ := LoadConfig(path)

	myconfig := Config{
		pd_user_token: config.pd_user_token,
		teams:         []string{"P8ODN2F"},
		silentUser:    "P8QS6CC",
		ignoredUsers:  []string{"PWT7CSF"},
	}
	ctx := context.Background()

	c, err := pd.NewConfig(myconfig.pd_user_token, myconfig.teams, myconfig.silentUser, myconfig.ignoredUsers)
	if err != nil {
		fmt.Println(err)
		return
	}
	users := difference(c.TeamsMemberIDs, myconfig.ignoredUsers)
	currentUser, _ := c.Client.GetUserWithContext(ctx, c.CurrentUser.ID, pagerduty.GetUserOptions{})

	fmt.Printf("%v", currentUser.Name)

	highAcknowledgedIncidents, err := c.Client.ListIncidentsWithContext(ctx, pagerduty.ListIncidentsOptions{UserIDs: users, Statuses: []string{"acknowledged"}, Urgencies: []string{"high"}})
	if err != nil {
		fmt.Println(err)
		return
	} else {
		for _, inc := range highAcknowledgedIncidents.Incidents {
			fmt.Printf("%v\n", inc)
		}

	}

	// triggered_incidents, err := c.Client.GetCurrentUserWithContext(ctx, pagerduty.GetCurrentUserOptions{})
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// fmt.Printf("%+v", triggered_incidents)

}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("osdctl")
	viper.SetConfigType("yaml")

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		return config, err
	}
	config.pd_user_token = viper.GetString("pd_user_token")

	return config, nil
}

func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
