package client

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// CreateMappingRule - Create API for Mapping Rule endpoint
func (c *ThreeScaleClient) CreateMappingRule(
	accessToken string, svcId string, method string,
	pattern string, delta int, metricId string) (MappingRuleResp, error) {

	var apiResp MappingRuleResp
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
		return apiResp, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return apiResp, genRespErr("mapping rule create", err.Error())
	}

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("mapping rule create", err.Error())
	}
	return apiResp, nil
}

func genMrEp(svcId string) string {
	return fmt.Sprintf(createMappingRuleEndpoint, svcId)
}
