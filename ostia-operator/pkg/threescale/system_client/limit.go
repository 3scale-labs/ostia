package client

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// CreateLimit - Adds a limit to a metric of an application plan.
// All applications with the application plan (application_plan_id) will be constrained by this new limit on the metric (metric_id).
func (c *ThreeScaleClient) CreateLimit(accessToken string, appPlanId string, metricId string, period string, value int) (Limit, error) {
	var apiResp Limit
	if value < 1 {
		return apiResp, errors.New("value must be positive")
	}

	ep := genCreateLimitEp(appPlanId, metricId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("application_plan_id", appPlanId)
	values.Add("metric_id", metricId)
	values.Add("period", period)
	values.Add("value", strconv.Itoa(value))

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return apiResp, httpReqError
	}
	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return apiResp, genRespErr("create limit", err.Error())
	}

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("create limit", err.Error())
	}
	return apiResp, nil

}

// ListLimits - Returns the list of all limits associated to an application plan.
func (c *ThreeScaleClient) ListLimitsPerPlan(accessToken string, appPlanId string) (LimitList, error) {
	return c.listLimits(genListLimitPerAppEp(appPlanId), accessToken)
}

// ListLimits - Returns the list of all limits associated to a metric of an application plan
func (c *ThreeScaleClient) ListLimitsPerMetric(accessToken string, appPlanId string, metricId string) (LimitList, error) {
	return c.listLimits(genListLimitPerMetricEp(appPlanId, metricId), accessToken)
}

// listLimits takes an endpoint and returns a list of limits
func (c *ThreeScaleClient) listLimits(ep string, accessToken string) (LimitList, error) {
	var ml LimitList

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
		return ml, genRespErr("list limits per plan", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return ml, genRespErr("list limits per plan", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&ml); err != nil {
		return ml, genRespErr("list limits per plan", err.Error())
	}
	return ml, nil
}

func genListLimitPerAppEp(appPlanId string) string {
	return fmt.Sprintf(listLimitPerAppPlanEndpoint, appPlanId)
}

func genListLimitPerMetricEp(appPlanId string, metric string) string {
	return fmt.Sprintf(listLimitPerMetricEndpoint, appPlanId, metric)
}

func genCreateLimitEp(appPlanId string, metricId string) string {
	return fmt.Sprintf(limitEndpoint, appPlanId, metricId)
}
