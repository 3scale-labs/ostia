package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CreateMappingRule - Create API for Mapping Rule endpoint
func (c *ThreeScaleClient) CreateMappingRule(
	accessToken string, svcId string, method string,
	pattern string, delta int, metricId string) (MappingRule, error) {

	var mr MappingRule
	ep := genMrEp(svcId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)
	values.Add("http_method", method)
	values.Add("pattern", pattern)
	values.Add("delta", strconv.Itoa(delta))
	values.Add("metric_id", metricId)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return mr, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return mr, genRespErr("mapping rule create", err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		return mr, genRespErr("mapping rule list", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return mr, genRespErr("mapping rule create", err.Error())
	}
	return mr, nil
}

// UpdateMetric - Updates a Proxy Mapping Rule
// The proxy object must be updated after a mapping rule update to apply the change to proxy config
// Valid params keys and their purpose are as follows:
// "http_method" - HTTP method
// "pattern"     - Mapping Rule pattern
// "delta"       - Increase the metric by this delta
// "metric_id"   - The metric ID
func (c *ThreeScaleClient) UpdateMappingRule(accessToken string, svcId string, id string, params Params) (MappingRule, error) {
	var m MappingRule

	ep := genMrUpdateEp(svcId, id)

	values := url.Values{}
	values.Add("access_token", accessToken)
	for k, v := range params {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(ep, body)
	if err != nil {
		return m, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return m, genRespErr("update mapping rule", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return m, genRespErr("update mapping rule", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&m); err != nil {
		return m, genRespErr("update mapping rule", err.Error())
	}
	return m, nil
}

// DeleteMappingRule - Deletes a Proxy Mapping Rule.
// The proxy object must be updated after a mapping rule deletion to apply the change to proxy config
func (c *ThreeScaleClient) DeleteMappingRule(accessToken string, svcId string, id string) error {
	ep := genMrUpdateEp(svcId, id)

	values := url.Values{}
	values.Add("access_token", accessToken)

	body := strings.NewReader(values.Encode())
	req, err := c.buildDeleteReq(ep, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return genRespErr("delete mapping rule", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return genRespErr("delete mapping rule", handleErrResp(resp))
	}
	return nil
}

// ListMappingRule - List API for Mapping Rule endpoint
func (c *ThreeScaleClient) ListMappingRule(accessToken string, svcId string) (MappingRuleList, error) {
	var mrl MappingRuleList
	ep := genMrEp(svcId)

	req, err := c.buildGetReq(ep)
	if err != nil {
		return mrl, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)

	req.URL.RawQuery = values.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return mrl, genRespErr("mapping rule list failed", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return mrl, genRespErr("mapping rule list failed", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&mrl); err != nil {
		return mrl, genRespErr("mapping rule list", err.Error())
	}

	return mrl, nil
}

func genMrEp(svcId string) string {
	return fmt.Sprintf(mappingRuleEndpoint, svcId)
}

func genMrUpdateEp(svcId string, id string) string {
	return fmt.Sprintf(updateDeleteMappingRuleEndpoint, svcId, id)
}
