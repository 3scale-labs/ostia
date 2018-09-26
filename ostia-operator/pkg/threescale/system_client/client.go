package client

// This package provides bare minimum functionality for all the endpoints it exposes,
// which is a subset of the Account Management API.
// No optional parameters can currently be provided.
// TODO - This is PoC quality code and is untested. Should not be merged back to master
// TODO without being DRY'ed  if possible and hardened significantly

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	createAppEndpoint           = "/admin/api/accounts/%s/applications.xml"
	createAppPlanEndpoint       = "/admin/api/services/%s/application_plans.xml"
	mappingRuleEndpoint         = "/admin/api/services/%s/proxy/mapping_rules.xml"
	createListMetricEndpoint    = "/admin/api/services/%s/metrics.xml"
	updateDeleteMetricEndpoint  = "/admin/api/services/%s/metrics/%s.xml"
	ListAppPlansByService       = "/admin/api/services/%s/application_plans.xml"
	ListAppPlans                = "/admin/api/application_plans.xml"
	AppPlanServiceEndpoint      = "/admin/api/services/%s/application_plans/%s.xml"
	limitEndpoint               = "/admin/api/application_plans/%s/metrics/%s/limits.xml"
	listLimitPerAppPlanEndpoint = "/admin/api/application_plans/%s/limits.xml"
	listLimitPerMetricEndpoint  = "/admin/api/application_plans/%s/metrics/%s/limits.xml"
	ServicesEndpoint            = "/admin/api/services.xml"
)

var httpReqError = errors.New("error building http request")

// Returns a custom AdminPortal which integrates with the users Account Management API.
// Supported schemes are http and https
func NewAdminPortal(scheme string, host string, port int) (*AdminPortal, error) {
	url2, err := verifyUrl(fmt.Sprintf("%s://%s:%d", scheme, host, port))
	if err != nil {
		return nil, err
	}
	return &AdminPortal{scheme, host, port, url2}, nil
}

// Creates a ThreeScaleClient to communicate with Account Management API.
// If http Client is nil, the default http client will be used
func NewThreeScale(backEnd *AdminPortal, httpClient *http.Client) *ThreeScaleClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ThreeScaleClient{backEnd, httpClient}
}

func NewParams() Params {
	params := make(map[string]string)
	return params
}

func (p Params) AddParam(key string, value string) {
	p[key] = value
}

// Request builder for GET request to the provided endpoint
func (c *ThreeScaleClient) buildGetReq(ep string) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("GET", c.adminPortal.baseUrl.ResolveReference(path).String(), nil)
	req.Header.Set("Accept", "application/xml")
	return req, err
}

// Request builder for POST request to the provided endpoint
func (c *ThreeScaleClient) buildPostReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("POST", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

// Request builder for PUT request to the provided endpoint
func (c *ThreeScaleClient) buildUpdateReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("PUT", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

// Request builder for DELETE request to the provided endpoint
func (c *ThreeScaleClient) buildDeleteReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("DELETE", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

// Request builder for PUT request to the provided endpoint
func (c *ThreeScaleClient) buildPutReq(ep string, body io.Reader) (*http.Request, error) {
	path := &url.URL{Path: ep}
	req, err := http.NewRequest("PUT", c.adminPortal.baseUrl.ResolveReference(path).String(), body)
	req.Header.Set("Accept", "application/xml")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, err
}

// Verifies a custom admin portal is valid
func verifyUrl(urlToCheck string) (*url.URL, error) {
	url2, err := url.ParseRequestURI(urlToCheck)
	if err == nil {
		if url2.Scheme != "http" && url2.Scheme != "https" {
			err = fmt.Errorf("unsupported schema %s passed to adminPortal", url2.Scheme)
		}

	}
	return url2, err
}

// Decodes and transforms an API response error into a string
func handleErrResp(resp *http.Response) string {
	var errResp ErrorResp
	errMsg := fmt.Sprintf("status code: %v", resp.StatusCode)
	err := xml.NewDecoder(resp.Body).Decode(&errResp)
	if err == nil {
		errMsg = fmt.Sprintf("%s - reason: %s", errMsg, errResp.Error.Text)
	}
	return errMsg
}

// Helper method to generate error message for client functions
func genRespErr(ep string, err string) error {
	return fmt.Errorf("error calling %s endpoint - %s", ep, err)
}
