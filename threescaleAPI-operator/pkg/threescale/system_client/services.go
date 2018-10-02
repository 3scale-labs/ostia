package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	serviceCreateList   = "/admin/api/services.xml"
	serviceUpdateDelete = "/admin/api/services/%s.xml"
)

func (c *ThreeScaleClient) CreateService(accessToken string, name string) (Service, error) {
	var s Service

	endpoint := serviceCreateList
	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("name", name)
	values.Add("system_name", name)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return s, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return s, genRespErr("create service", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return s, genRespErr("create service", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&s); err != nil {
		return s, genRespErr("create service", err.Error())
	}
	return s, nil
}

// UpdateService - Update the service. Valid params keys and their purpose are as follows:
// "name"                - Name of the service.
// "support_email"       - New support email.
// "tech_support_email"  - New tech support email.
// "admin_support_email" - New admin support email.
// "deployment_option"   - Deployment option for the gateway: 'hosted' for APIcast hosted, 'self-managed' for APIcast Self-managed option
// "backend_version"     - Authentication mode: '1' for API key, '2' for App Id / App Key, 'oauth' for OAuth mode, 'oidc' for OpenID Connect
func (c *ThreeScaleClient) UpdateService(accessToken string, id string, params Params) (Service, error) {
	var s Service

	endpoint := fmt.Sprintf(serviceUpdateDelete, id)

	values := url.Values{}
	values.Add("access_token", accessToken)
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(endpoint, body)
	if err != nil {
		return s, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return s, genRespErr("update service", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return s, genRespErr("update service", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&s); err != nil {
		return s, genRespErr("update service", err.Error())
	}
	return s, nil
}

// DeleteService - Delete the service.
// Deleting a service removes all applications and service subscriptions.
func (c *ThreeScaleClient) DeleteService(accessToken string, id string) error {
	endpoint := fmt.Sprintf(serviceUpdateDelete, id)

	values := url.Values{}
	values.Add("access_token", accessToken)

	body := strings.NewReader(values.Encode())
	req, err := c.buildDeleteReq(endpoint, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return genRespErr("delete service", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return genRespErr("delete service", handleErrResp(resp))
	}
	return nil
}

func (c *ThreeScaleClient) ListServices(accessToken string) (ServiceList, error) {
	var sl ServiceList

	ep := serviceCreateList

	req, err := c.buildGetReq(ep)
	if err != nil {
		return sl, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return sl, genRespErr("List Services:", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return sl, genRespErr("List Services:", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&sl); err != nil {
		return sl, genRespErr("List Services:", err.Error())
	}

	return sl, nil
}
