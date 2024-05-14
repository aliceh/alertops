package main

import (
	"context"
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
)

var authtoken = "" // Set your auth token here

func main() {
	ctx := context.Background()
	client := pagerduty.NewClient(authtoken)

	var opts pagerduty.ListEscalationPoliciesOptions
	eps, err := client.ListEscalationPoliciesWithContext(ctx, opts)
	if err != nil {
		panic(err)
	}
	for _, p := range eps.EscalationPolicies {
		fmt.Println(p.Name)
	}
}
