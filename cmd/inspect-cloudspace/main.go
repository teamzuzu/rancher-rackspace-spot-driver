// inspect-cloudspace fetches a named CloudSpace from the Rackspace Spot API
// and prints its raw JSON spec. Use this to discover what field names and
// values the Spot platform uses for HA clusters.
//
// Usage:
//
//	go run ./cmd/inspect-cloudspace \
//	  --refresh-token <token> \
//	  --org <org-name> \
//	  --cloudspace <cloudspace-name>
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	spotv1 "github.com/rackspace-spot/spot-go-sdk/api/v1"
)

func main() {
	refreshToken := flag.String("refresh-token", os.Getenv("SPOT_REFRESH_TOKEN"), "Rackspace Spot refresh token")
	org := flag.String("org", "", "Organisation name or ID")
	cloudspace := flag.String("cloudspace", "", "CloudSpace name to inspect (leave blank to list all)")
	flag.Parse()

	if *refreshToken == "" {
		fmt.Fprintln(os.Stderr, "error: --refresh-token is required (or set SPOT_REFRESH_TOKEN)")
		os.Exit(1)
	}
	if *org == "" {
		fmt.Fprintln(os.Stderr, "error: --org is required")
		os.Exit(1)
	}

	ctx := context.Background()

	c, err := spotv1.NewSpotClient(&spotv1.Config{RefreshToken: *refreshToken})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating client: %v\n", err)
		os.Exit(1)
	}
	if _, err := c.Authenticate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "error authenticating: %v\n", err)
		os.Exit(1)
	}

	_, orgID, err := c.GetOrgID(ctx, *org)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving org %q: %v\n", *org, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "org=%q orgID=%q baseURL=%q\n", *org, orgID, c.BaseURL)

	var url string
	if *cloudspace == "" {
		url = fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces", c.BaseURL, orgID)
	} else {
		url = fmt.Sprintf("%s/apis/ngpc.rxt.io/v1/namespaces/%s/cloudspaces/%s", c.BaseURL, orgID, *cloudspace)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Fprintf(os.Stderr, "HTTP %d\n", resp.StatusCode)

	// Pretty-print if JSON.
	var pretty interface{}
	if json.Unmarshal(body, &pretty) == nil {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(pretty)
	} else {
		fmt.Println(string(body))
	}
}
