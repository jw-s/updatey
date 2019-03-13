package docker

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
)

/*
The code below is from github.com/docker/docker/registry/auth.go
This is licensed under Apache-2.0
*/

// DefaultRegistryVersionHeader is the name of the default HTTP header
// that carries Registry version info
var DefaultRegistryVersionHeader = "Docker-Distribution-Api-Version"

// PingResponseError is used when the response from a ping
// was received but invalid.
type PingResponseError struct {
	Err error
}

func (err PingResponseError) Error() string {
	return err.Err.Error()
}

// PingV2Registry attempts to ping a v2 registry and on success return a
// challenge manager for the supported authentication types and
// whether v2 was confirmed by the response. If a response is received but
// cannot be interpreted a PingResponseError will be returned.
func PingV2Registry(endpoint *url.URL, transport http.RoundTripper) (challenge.Manager, bool, error) {
	var (
		foundV2   = false
		v2Version = auth.APIVersion{
			Type:    "registry",
			Version: "2.0",
		}
	)

	pingClient := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}
	endpointStr := strings.TrimRight(endpoint.String(), "/") + "/v2/"
	req, err := http.NewRequest("GET", endpointStr, nil)
	if err != nil {
		return nil, false, err
	}
	resp, err := pingClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	versions := auth.APIVersions(resp, DefaultRegistryVersionHeader)
	for _, pingVersion := range versions {
		if pingVersion == v2Version {
			// The version header indicates we're definitely
			// talking to a v2 registry. So don't allow future
			// fallbacks to the v1 protocol.

			foundV2 = true
			break
		}
	}

	challengeManager := challenge.NewSimpleManager()
	if err := challengeManager.AddResponse(resp); err != nil {
		return nil, foundV2, PingResponseError{
			Err: err,
		}
	}

	return challengeManager, foundV2, nil
}
