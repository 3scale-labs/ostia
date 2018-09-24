package client

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

// CreateMetric - Creates a metric on a service. All metrics are scoped by service.
func (c *ThreeScaleClient) CreateMetric(accessToken string, svcId string, name string, unit string) (MetricResp, error) {
	var apiResp MetricResp

	ep := genMetricEp(svcId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)
	values.Add("friendly_name", name)
	values.Add("unit", unit)

	body := strings.NewReader(values.Encode())

	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return apiResp, httpReqError
	}
	resp, err := c.httpClient.Do(req)

	defer resp.Body.Close()

	if err != nil {
		return apiResp, genRespErr("create metric", err.Error())
	}

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("create metric", err.Error())
	}
	return apiResp, nil
}

func genMetricEp(svcID string) string {
	return fmt.Sprintf(createMetricEndpoint, svcID)
}
