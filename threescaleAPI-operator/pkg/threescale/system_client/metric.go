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

	ep := genMetricCreateListEp(svcId)

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
	if err != nil {
		return m, genRespErr("create metric", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return m, genRespErr("create metric", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&m); err != nil {
		return m, genRespErr("create metric", err.Error())
	}
	return m, nil
}

// UpdateMetric - Updates the metric of a service. Valid params keys and their purpose are as follows:
// "friendly_name" - Name of the metric.
// "unit" - Measure unit of the metric.
// "description" - Description of the metric.
func (c *ThreeScaleClient) UpdateMetric(accessToken string, svcId string, id string, params Params) (Metric, error) {
	var m Metric

	ep := genMetricUpdateDeleteEp(svcId, id)

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
	if err != nil {
		return m, genRespErr("update metric", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return m, genRespErr("update metric", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&m); err != nil {
		return m, genRespErr("update metric", err.Error())
	}
	return m, nil
}

// DeleteMetric - Deletes the metric of a service.
// When a metric is deleted, the associated limits across application plans are removed
func (c *ThreeScaleClient) DeleteMetric(accessToken string, svcId string, id string) error {
	ep := genMetricUpdateDeleteEp(svcId, id)

	values := url.Values{}
	values.Add("access_token", accessToken)

	body := strings.NewReader(values.Encode())
	req, err := c.buildDeleteReq(ep, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return genRespErr("delete metric", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return genRespErr("update metric", handleErrResp(resp))
	}
	return nil
}

// ListMetric - Returns the list of metrics of a service
func (c *ThreeScaleClient) ListMetrics(accessToken string, svcId string) (MetricList, error) {
	var ml MetricList

	ep := genMetricCreateListEp(svcId)

	req, err := c.buildGetReq(ep)
	if err != nil {
		return ml, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return ml, genRespErr("create metric", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ml, genRespErr("create metric", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&ml); err != nil {
		return ml, genRespErr("create metric", err.Error())
	}
	return ml, nil
}

func genMetricCreateListEp(svcID string) string {
	return fmt.Sprintf(createListMetricEndpoint, svcID)
}

func genMetricUpdateDeleteEp(svcID string, metricId string) string {
	return fmt.Sprintf(updateDeleteMetricEndpoint, svcID, metricId)
}
