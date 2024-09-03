package pd

import (
	"context"
	"fmt"
	"strings"

	"github.com/PagerDuty/go-pagerduty"
	utils "github.com/aliceh/alertops/pkg/utils"
)

const (
	defaultPageLimit = 100
	defaultOffset    = 0
)

type Alert struct {
	IncidentID  string
	AlertID     string
	ClusterID   string
	ClusterName string
	Name        string
	Console     string
	Hostname    string
	IP          string
	Labels      string
	LastCheckIn string
	Severity    string
	Status      string
	Sop         string
	Token       string
	Tags        string
	WebURL      string
}

var defaultIncidentStatues = []string{"triggered", "acknowledged"}

// PagerDutyClientInterface is an interface that defines the methods used by the pd package and makes it easier to mock
// calls to PagerDuty in tests
type PagerDutyClientInterface interface {
	CreateIncidentNoteWithContext(ctx context.Context, id string, note pagerduty.IncidentNote) (*pagerduty.IncidentNote, error)
	GetCurrentUserWithContext(ctx context.Context, opts pagerduty.GetCurrentUserOptions) (*pagerduty.User, error)
	GetIncidentWithContext(ctx context.Context, id string) (*pagerduty.Incident, error)
	GetService(serviceID string, opts *pagerduty.GetServiceOptions) (*pagerduty.Service, error)
	GetTeamWithContext(ctx context.Context, id string) (*pagerduty.Team, error)
	ListMembersWithContext(ctx context.Context, id string, opts pagerduty.ListTeamMembersOptions) (*pagerduty.ListTeamMembersResponse, error)
	GetUserWithContext(ctx context.Context, id string, opts pagerduty.GetUserOptions) (*pagerduty.User, error)
	ListIncidentAlertsWithContext(ctx context.Context, id string, opts pagerduty.ListIncidentAlertsOptions) (*pagerduty.ListAlertsResponse, error)
	ListIncidentsWithContext(ctx context.Context, opts pagerduty.ListIncidentsOptions) (*pagerduty.ListIncidentsResponse, error)
	ListIncidentNotesWithContext(ctx context.Context, id string) ([]pagerduty.IncidentNote, error)
	ManageIncidentsWithContext(ctx context.Context, email string, opts []pagerduty.ManageIncidentsOptions) (*pagerduty.ListIncidentsResponse, error)
}

// GetClusterName interacts with the PD service endpoint and returns the cluster name string.
func GetClusterName(servideID string, c PagerDutyClient) (string, error) {
	service, err := c.GetService(servideID, &pagerduty.GetServiceOptions{})

	if err != nil {
		return "", err
	}

	clusterName := strings.Split(service.Description, " ")[0]

	return clusterName, nil
}

// ParseAlertData parses a pagerduty alert data into the Alert struct.
func (a *Alert) ParseAlertData(c PagerDutyClient, alert *pagerduty.IncidentAlert) (err error) {
	a.IncidentID = alert.Incident.ID
	a.AlertID = alert.ID
	a.Name = alert.Summary
	a.Status = alert.Status
	a.WebURL = alert.HTMLURL

	// Check if the alert is of type 'Missing cluster'
	isCHGM := alert.Body["details"].(map[string]interface{})["notes"]

	// Check if the alert is of type 'Certificate is expiring'
	isCertExpiring := alert.Body["details"].(map[string]interface{})["hostname"]

	if isCHGM != nil {
		notes := strings.Split(fmt.Sprint(alert.Body["details"].(map[string]interface{})["notes"]), "\n")

		a.ClusterID = strings.Replace(notes[0], "cluster_id: ", "", 1)
		a.ClusterName = strings.Split(fmt.Sprint(alert.Body["details"].(map[string]interface{})["name"]), ".")[0]

		lastCheckIn := fmt.Sprint(alert.Body["details"].(map[string]interface{})["last healthy check-in"])
		a.LastCheckIn, err = utils.FormatTimestamp(lastCheckIn)

		if err != nil {
			return err
		}

		a.Token = fmt.Sprint(alert.Body["details"].(map[string]interface{})["token"])
		a.Tags = fmt.Sprint(alert.Body["details"].(map[string]interface{})["tags"])
		a.Sop = strings.Replace(notes[1], "runbook: ", "", 1)

	} else if isCertExpiring != nil {
		a.Hostname = fmt.Sprint(alert.Body["details"].(map[string]interface{})["hostname"])
		a.IP = fmt.Sprint(alert.Body["details"].(map[string]interface{})["ip"])
		a.Sop = fmt.Sprint(alert.Body["details"].(map[string]interface{})["url"])
		a.Name = strings.Split(alert.Summary, " on ")[0]
		a.ClusterName = "N/A"

	} else {
		a.ClusterID = fmt.Sprint(alert.Body["details"].(map[string]interface{})["cluster_id"])
		a.ClusterName, err = GetClusterName(alert.Service.ID, c)

		// If the service mapped to the current incident is not available (404)
		if err != nil {
			a.ClusterName = "N/A"
		}

		a.Console = fmt.Sprint(alert.Body["details"].(map[string]interface{})["console"])
		a.Labels = fmt.Sprint(alert.Body["details"].(map[string]interface{})["firing"])
		a.Sop = fmt.Sprint(alert.Body["details"].(map[string]interface{})["link"])
	}

	// If there's no cluster ID related to the given alert
	if a.ClusterID == "" {
		a.ClusterID = "N/A"
	}

	return nil
}

// PagerDutyClient implements PagerDutyClientInterface and is used by the pd package to make calls to PagerDuty
// This allows for mocking calls that would usually use the pagerduty.Client struct
type PagerDutyClient interface {
	PagerDutyClientInterface
}

// Config is a struct that holds the PagerDuty client used for all the PagerDuty calls, and the config info for
// teams, silent user, and ignored users
type Config struct {
	Client      *pagerduty.Client
	CurrentUser *pagerduty.User

	// List of the users in the Teams
	TeamsMemberIDs []string
	Teams          []*pagerduty.Team

	SilentUser   *pagerduty.User
	IgnoredUsers []*pagerduty.User
}

func NewConfig(token string, teams []string, silentUser string, ignoredUsers []string) (*Config, error) {
	var c Config
	var err error

	c.Client = newClient(token)

	c.CurrentUser, err = c.Client.GetCurrentUserWithContext(context.Background(), pagerduty.GetCurrentUserOptions{})
	if err != nil {
		return &c, fmt.Errorf("pd.NewConfig(): failed to retrieve PagerDuty user: %v", err)
	}

	c.Teams, err = GetTeams(c.Client, teams)
	if err != nil {
		return &c, fmt.Errorf("pd.NewConfig(): failed to get team(s) `%v`: %v", teams, err)
	}

	c.TeamsMemberIDs, err = GetTeamMemberIDs(c.Client, c.Teams, pagerduty.ListTeamMembersOptions{Limit: defaultPageLimit, Offset: defaultOffset})
	if err != nil {
		return &c, fmt.Errorf("pd.NewConfig(): failed to get users(s) from teams: %v", err)
	}

	c.SilentUser, err = GetUser(c.Client, silentUser, pagerduty.GetUserOptions{})
	if err != nil {
		return &c, fmt.Errorf("pd.NewConfig(): failed to get silent user: %v", err)
	}

	for _, i := range ignoredUsers {
		user, err := GetUser(c.Client, i, pagerduty.GetUserOptions{})
		if err != nil {
			return &c, fmt.Errorf("pd.NewConfig(): failed to get user for ignore list `%v`: %v", i, err)
		}
		c.IgnoredUsers = append(c.IgnoredUsers, user)
	}

	return &c, nil
}

func newClient(token string) *pagerduty.Client {
	return pagerduty.NewClient(token)
}

func NewListIncidentOptsFromDefaults() pagerduty.ListIncidentsOptions {
	return pagerduty.ListIncidentsOptions{
		Limit:    defaultPageLimit,
		Offset:   defaultOffset,
		Statuses: defaultIncidentStatues,
	}

}

func HighAcknowledgedIncidents(client PagerDutyClient, users []string) (*pagerduty.ListIncidentsResponse, error) {
	highAcknowledgedIncidents, err := client.ListIncidentsWithContext(context.TODO(), pagerduty.ListIncidentsOptions{UserIDs: users, Urgencies: []string{"high"}, Statuses: []string{"acknowledged"}})

	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		for _, inc := range highAcknowledgedIncidents.Incidents {
			ack := inc.Acknowledgements
			id := inc.ID
			description := inc.Description
			cluster := inc.HTMLURL
			service := inc.Service
			fmt.Printf("\n")
			fmt.Printf("ACK: %v\n", ack)
			fmt.Printf("ID: %v\n", id)
			fmt.Printf("Description: %v\n", description)
			fmt.Printf("Cluster: %v\n", cluster)
			fmt.Printf("Service: %v\n", service)
			fmt.Printf("\n")
		}

	}
	return highAcknowledgedIncidents, err

}

func AcknowledgeIncident(client PagerDutyClient, incidents []*pagerduty.Incident, user *pagerduty.User) ([]pagerduty.Incident, error) {
	var i []pagerduty.Incident

	opts := []pagerduty.ManageIncidentsOptions{}

	for _, incident := range incidents {
		opts = append(opts, pagerduty.ManageIncidentsOptions{
			ID:     incident.ID,
			Status: "acknowledged",
			Assignments: []pagerduty.Assignee{{
				Assignee: user.APIObject,
			}},
		})
	}

	for {
		response, err := client.ManageIncidentsWithContext(context.TODO(), user.Email, opts)
		if err != nil {
			return i, fmt.Errorf("pd.AcknowledgeIncident(): failed to acknowledge incident(s) `%v`: %v", incidents, err)
		}

		i = append(i, response.Incidents...)

		if response.More {
			panic("pd.AcknowledgeIncident(): PagerDuty response indicated more data available")
		}

		if !response.More {
			break
		}

	}

	return i, nil
}

func GetAlerts(client PagerDutyClient, id string, opts pagerduty.ListIncidentAlertsOptions) ([]pagerduty.IncidentAlert, error) {
	var a []pagerduty.IncidentAlert

	for {
		response, err := client.ListIncidentAlertsWithContext(context.TODO(), id, opts)
		if err != nil {
			return a, fmt.Errorf("pd.GetAlerts(): failed to get alerts for incident `%v`: %v", id, err)
		}

		a = append(a, response.Alerts...)

		opts.Offset += opts.Limit

		if !response.More {
			break
		}
	}

	return a, nil
}

func GetIncident(client PagerDutyClient, id string) (*pagerduty.Incident, error) {
	var i *pagerduty.Incident

	i, err := client.GetIncidentWithContext(context.TODO(), id)
	if err != nil {
		return i, fmt.Errorf("pd.GetIncident(): failed to get incident `%v`: %v", id, err)
	}

	return i, nil
}

func GetIncidents(client PagerDutyClient, opts pagerduty.ListIncidentsOptions) ([]pagerduty.Incident, error) {
	var i []pagerduty.Incident

	for {
		response, err := client.ListIncidentsWithContext(context.TODO(), opts)
		if err != nil {
			return i, fmt.Errorf("pd.GetIncidents(): failed to get incidents : %v", err)
		}

		i = append(i, response.Incidents...)

		opts.Offset += opts.Limit

		if !response.More {
			break
		}
	}

	return i, nil
}

func GetNotes(client PagerDutyClient, id string) ([]pagerduty.IncidentNote, error) {
	var n []pagerduty.IncidentNote

	n, err := client.ListIncidentNotesWithContext(context.TODO(), id)
	if err != nil {
		return n, fmt.Errorf("pd.GetNotes(): failed to get incident notes `%v`: %v", id, err)
	}

	return n, nil
}

func GetTeams(client *pagerduty.Client, teams []string) ([]*pagerduty.Team, error) {
	var ctx = context.Background()
	var t []*pagerduty.Team

	for _, i := range teams {
		team, err := client.GetTeamWithContext(ctx, i)
		if err != nil {
			return t, fmt.Errorf("pd.GetTeams(): failed to find PagerDuty team `%v`: %v", i, err)
		}
		t = append(t, team)
	}

	return t, nil
}

func GetTeamMemberIDs(client *pagerduty.Client, teams []*pagerduty.Team, opts pagerduty.ListTeamMembersOptions) ([]string, error) {
	var ctx = context.Background()
	var u []string

	for _, team := range teams {
		for {
			response, err := client.ListMembersWithContext(ctx, team.ID, opts)
			if err != nil {
				return u, fmt.Errorf("pd.GetUsers(): failed to retrieve users for PagerDuty team `%v`: %v", team.ID, err)
			}

			for _, member := range response.Members {
				u = append(u, member.User.ID)
			}

			opts.Offset += opts.Limit

			if !response.More {
				break
			}
		}
	}

	return u, nil
}

func GetUser(client *pagerduty.Client, id string, opts pagerduty.GetUserOptions) (*pagerduty.User, error) {
	var ctx = context.Background()
	var u *pagerduty.User

	u, err := client.GetUserWithContext(ctx, id, opts)
	if err != nil {
		return u, fmt.Errorf("pd.GetUser(): failed to find PagerDuty user `%v`: %v", id, err)
	}

	return u, nil
}

func ReassignIncidents(client PagerDutyClient, incidents []*pagerduty.Incident, user *pagerduty.User, users []*pagerduty.User) ([]pagerduty.Incident, error) {
	var i []pagerduty.Incident

	a := []pagerduty.Assignee{}
	for _, user := range users {
		a = append(a, pagerduty.Assignee{Assignee: user.APIObject})
	}

	opts := []pagerduty.ManageIncidentsOptions{}

	for _, incident := range incidents {
		if incident == nil {
			return i, fmt.Errorf("pd.ReassignIncidents(): incident is nil")
		}
		opts = append(opts, pagerduty.ManageIncidentsOptions{
			ID:          incident.ID,
			Assignments: a,
		})
	}

	// This loop is likely unnecessary, as the "More" response is probably not used by PagerDuty here
	// but I'm including it in case we need to use it in the future, and raising a panic if we receive
	// a "More" response so we can fix the code

	for {
		response, err := client.ManageIncidentsWithContext(context.TODO(), user.Email, opts)
		if err != nil {
			return i, err
		}

		if response.More {
			// If we ever do get a "More" response, we we need to handle it, so panic to call attention to the problem
			panic("pd.ReassignIncidents(): PagerDuty response indicated more data available")
		}

		i = append(i, response.Incidents...)

		if !response.More {
			break
		}
	}

	return i, nil
}

func PostNote(client PagerDutyClient, id string, user *pagerduty.User, content string) (*pagerduty.IncidentNote, error) {
	var n *pagerduty.IncidentNote

	note := pagerduty.IncidentNote{
		Content: content,
		User:    user.APIObject,
	}

	n, err := client.CreateIncidentNoteWithContext(context.TODO(), id, note)
	if err != nil {
		return n, err
	}

	return n, nil
}
