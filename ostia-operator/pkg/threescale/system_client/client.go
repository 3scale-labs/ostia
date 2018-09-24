package client

// This package provides bare minimum functionality for all the endpoints it exposes.
// No optional parameters can currently be provided.
// TODO - This is PoC quality code and is untested. Should not be merged back to master
// TODO without being DRY'ed  if possible and hardened significantly

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	createAppEndpoint         = "/admin/api/accounts/%s/applications.xml"
	createAppPlanEndpoint     = "/admin/api/services/%s/application_plans.xml"
	createLimitEndpoint       = "/admin/api/application_plans/%s/metrics/%s/limits.xml"
	createMappingRuleEndpoint = "/admin/api/services/%s/proxy/mapping_rules.xml"
	createMetricEndpoint      = "/admin/api/services/%s/metrics.xml"
)

var httpReqError = errors.New("error building http request")

// Returns a custom Backend
// Can be used for on-premise installations
// Supported schemes are http and https
func NewBackend(scheme string, host string, port int) (*Backend, error) {
	url2, err := verifyBackendUrl(fmt.Sprintf("%s://%s:%d", scheme, host, port))
	if err != nil {
		return nil, err
	}
	return &Backend{scheme, host, port, url2}, nil
}

// Creates a ThreeScaleClient to communicate with the provided backend.
// If http Client is nil, the default http client will be used
func NewThreeScale(backEnd *Backend, httpClient *http.Client) *ThreeScaleClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ThreeScaleClient{backEnd, httpClient}
}

// Request builder for GET request to the provided endpoint
func (c *ThreeScaleClient) buildPostReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("POST", c.backend.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

// Verifies a custom backend is valid
func verifyBackendUrl(urlToCheck string) (*url.URL, error) {
	url2, err := url.ParseRequestURI(urlToCheck)
	if err == nil {
		if url2.Scheme != "http" && url2.Scheme != "https" {
			err = fmt.Errorf("unsupported schema %s passed to backend", url2.Scheme)
		}

	}
	return url2, err
}

// Helper method to generate error message for client functions
func genRespErr(ep string, err string) error {
	return fmt.Errorf("error calling %s endpoint - %s", ep, err)
}
