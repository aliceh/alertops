package pd

import (
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	config "github.com/aliceh/alertops/pkg/config"
	utils "github.com/aliceh/alertops/pkg/utils"
)

func process_alert(conf config.Config) {
	c, err := NewConfig(conf.Token, conf.Teams, conf.SilentUser, conf.IgnoredUsers)

	var Users = utils.DifferenceOfSlices(c.TeamsMemberIDs, conf.IgnoredUsers)

	highAcknowledgedIncidents, _ := HighAcknowledgedIncidents(c.Client, Users)
	for _, inc := range highAcknowledgedIncidents.Incidents {
		id := inc.ID

		alerts, _ := GetAlerts(c.Client, id, pagerduty.ListIncidentAlertsOptions{})
		for _, alert := range alerts {

			details := alert.Body["details"].(map[string]interface{})
			cluster_id := details["cluster_id"].(string)
			fmt.Printf("\n CLUSTER ID:,%v ,\n", cluster_id)

		}
		fmt.Printf("\n,%v ,\n", alerts)
	}
	fmt.Printf("\n")
	fmt.Printf("END Testing GetAlerts command\n")

	a, err := GetIncidents(c.Client, pagerduty.ListIncidentsOptions{})
	fmt.Printf("GetIncidents Output:  %v\n %v\n", a, err)
}
