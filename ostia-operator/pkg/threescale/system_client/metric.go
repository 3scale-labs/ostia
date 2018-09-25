package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CreateMetric - Creates a metric on a service. All metrics are scoped by service.
func (c *ThreeScaleClient) CreateMetric(accessToken string, svcId string, name string, unit string) (Metric, error) {
	var m Metric

	ep := genMetricEp(svcId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)
	values.Add("friendly_name", name)
	values.Add("unit", unit)

	body := strings.NewReader(values.Encode())

	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return m, httpReqError
	}
	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return m, genRespErr("create metric", err.Error())
	}

	if resp.StatusCode != http.StatusCreated {
		return m, genRespErr("create metric", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&m); err != nil {
		return m, genRespErr("create metric", err.Error())
	}
	return m, nil
}

// ListMetric - Returns the list of metrics of a service
func (c *ThreeScaleClient) ListMetrics(accessToken string, svcId string) (MetricList, error) {
	var ml MetricList

	ep := genMetricEp(svcId)

	req, err := c.buildGetReq(ep)
	if err != nil {
		return ml, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return ml, genRespErr("create metric", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return ml, genRespErr("create metric", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&ml); err != nil {
		return ml, genRespErr("create metric", err.Error())
	}
	return ml, nil
}

func genMetricEp(svcID string) string {
	return fmt.Sprintf(metricEndpoint, svcID)
}
