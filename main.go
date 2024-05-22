package main

import (
	"fmt"

	pd "github.com/aliceh/alertops/pkg/provider"
)

var token = "u+ZoxVdhSbar1qxnHyvQ"

func main() {

	client := pd.NewClient().WithOauthToken(token)

	client.GetPDServiceIDs()
	fmt.Println("Hello world!")

}
