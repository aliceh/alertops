package main

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	config "github.com/aliceh/alertops/pkg/config"
	pd "github.com/aliceh/alertops/pkg/pagerduty"
	utils "github.com/aliceh/alertops/pkg/utils"
)

func main() {

	config, err := config.LoadConfig(config.Path)
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()

	c, err := pd.NewConfig(config.Token, config.Teams, config.SilentUser, config.IgnoredUsers)
	if err != nil {
		fmt.Println(err)
		return
	}
	users := utils.DifferenceOfSlices(c.TeamsMemberIDs, config.IgnoredUsers)
	currentUser, _ := c.Client.GetUserWithContext(ctx, c.CurrentUser.ID, pagerduty.GetUserOptions{})
	fmt.Printf("%v\n", currentUser.Name)

	highAcknowledgedIncidents, err := c.Client.ListIncidentsWithContext(ctx, pagerduty.ListIncidentsOptions{UserIDs: users, Urgencies: []string{"high"}, Statuses: []string{"acknowledged"}})

	if err != nil {
		fmt.Println(err)
		return
	} else {
		for _, inc := range highAcknowledgedIncidents.Incidents {
			ack := inc.Acknowledgements
			id := inc.ID
			description := inc.Description
			cluster := inc.HTMLURL
			service := inc.Service
			fmt.Printf("%v\n", ack)
			fmt.Printf("%v\n", id)
			fmt.Printf("%v\n", description)
			fmt.Printf("%v\n", cluster)
			fmt.Printf("%v\n", service)
		}

	}

	// triggered_incidents, err := c.Client.GetCurrentUserWithContext(ctx, pagerduty.GetCurrentUserOptions{})
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// fmt.Printf("%+v", triggered_incidents)

}
