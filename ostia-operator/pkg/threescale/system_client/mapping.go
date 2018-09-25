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
	defer resp.Body.Close()

	if err != nil {
		return mrl, genRespErr("mapping rule list", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return mrl, genRespErr("mapping rule list", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&mrl); err != nil {
		return mrl, genRespErr("mapping rule list", err.Error())
	}

	return mrl, nil
}

func genMrEp(svcId string) string {
	return fmt.Sprintf(mappingRuleEndpoint, svcId)
}
