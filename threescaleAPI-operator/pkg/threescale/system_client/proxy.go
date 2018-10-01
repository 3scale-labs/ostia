package client

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	proxyGetUpdate       = "/admin/api/services/%s/proxy.xml"
	proxyConfigGet       = "/admin/api/services/%s/proxy/configs/%s/%s.json"
	proxyConfigList      = "/admin/api/services/%s/proxy/configs/%s.json"
	proxyConfigLatestGet = "/admin/api/services/%s/proxy/configs/%s/latest.json"
	proxyConfigPromote   = "/admin/api/services/%s/proxy/configs/%s/%s/promote.json"
)

// ReadProxy - Returns the Proxy for a specific Service.
func (c *ThreeScaleClient) ReadProxy(accessToken string, svcID string) (Proxy, error) {
	var p Proxy

	endpoint := fmt.Sprintf(proxyGetUpdate, svcID)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return p, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return p, genRespErr("read proxy", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return p, genRespErr("read proxy", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&p); err != nil {
		return p, genRespErr("read proxy", err.Error())
	}
	return p, nil
}

// GetProxyConfig - Returns the Proxy Configs of a Service
func (c *ThreeScaleClient) GetProxyConfig(accessToken string, svcId string, env string, version string) (ProxyConfigElement, error) {
	endpoint := fmt.Sprintf(proxyConfigGet, svcId, env, version)
	return c.getProxyConfig(accessToken, endpoint)
}

// GetLatestProxyConfig - Returns the latest Proxy Config
func (c *ThreeScaleClient) GetLatestProxyConfig(accessToken string, svcId string, env string) (ProxyConfigElement, error) {
	endpoint := fmt.Sprintf(proxyConfigLatestGet, svcId, env)
	return c.getProxyConfig(accessToken, endpoint)
}

// UpdateProxy - Changes the Proxy settings.
// This will create a new APIcast configuration version for the Staging environment with the updated settings.
func (c *ThreeScaleClient) UpdateProxy(accessToken string, svcId string, params Params) (Proxy, error) {
	var p Proxy

	endpoint := fmt.Sprintf(proxyGetUpdate, svcId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(endpoint, body)
	if err != nil {
		return p, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return p, genRespErr("update proxy", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return p, genRespErr("update proxy", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&p); err != nil {
		return p, genRespErr("update proxy", err.Error())
	}
	return p, nil
}

// ListProxyConfig - Returns the Proxy Configs of a Service
// env parameter should be one of 'sandbox', 'production'
func (c *ThreeScaleClient) ListProxyConfig(accessToken string, svcId string, env string) (ProxyConfigList, error) {
	var pc ProxyConfigList

	endpoint := fmt.Sprintf(proxyConfigList, svcId, env)
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return pc, httpReqError
	}
	req.Header.Set("Accept", "application/json")

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return pc, genRespErr("list proxy configs", err.Error())
	}

	// TODO - Add some generic json handling
	if resp.StatusCode != http.StatusOK {
		return pc, fmt.Errorf("error calling list proxy config api - status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&pc)
	if err != nil {
		return pc, fmt.Errorf("error decoding list proxy config json response - %s", err)
	}

	return pc, nil
}

// PromoteProxyConfig - Promotes a Proxy Config from one environment to another environment.
func (c *ThreeScaleClient) PromoteProxyConfig(accessToken string, svcId string, env string, version string, toEnv string) (ProxyConfigElement, error) {
	var pe ProxyConfigElement
	endpoint := fmt.Sprintf(proxyConfigPromote, svcId, env, version)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("to", toEnv)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return pe, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return pe, genRespErr("proxy promote", err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		return pe, fmt.Errorf("error calling proxy promote api - status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&pe)
	if err != nil {
		return pe, fmt.Errorf("error decoding proxy config json response - %s", err)
	}

	return pe, nil
}

func (c *ThreeScaleClient) getProxyConfig(token, endpoint string) (ProxyConfigElement, error) {
	var pc ProxyConfigElement
	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return pc, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", token)
	req.URL.RawQuery = values.Encode()
	req.Header.Set("accept", "application/json")

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return pc, genRespErr("get latest proxy config", err.Error())
	}

	// TODO - Add some generic json handling
	if resp.StatusCode != http.StatusOK {
		return pc, fmt.Errorf("error calling get latest proxy config api - status %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&pc)
	if err != nil {
		return pc, fmt.Errorf("error decoding latest proxy config json response - %s", err)
	}

	return pc, nil
}
