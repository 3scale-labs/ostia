package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	limitAppPlanCreate           = "/admin/api/application_plans/%s/metrics/%s/limits.xml"
	limitAppPlanList             = "/admin/api/application_plans/%s/limits.xml"
	limitAppPlanUpdateDelete     = "/admin/api/application_plans/%s/metrics/%s/limits/%s.xml "
	limitAppPlanMetricList       = "/admin/api/application_plans/%s/metrics/%s/limits.xml"
	limitEndUserPlanCreateList   = "/admin/api/end_user_plans/%s/metrics/%s/limits.xml"
	limitEndUserPlanUpdateDelete = "/admin/api/end_user_plans/%s/metrics/%s/limits/%s.xml"
)

// CreateLimitAppPlan - Adds a limit to a metric of an application plan.
// All applications with the application plan (application_plan_id) will be constrained by this new limit on the metric (metric_id).
func (c *ThreeScaleClient) CreateLimitAppPlan(accessToken string, appPlanId string, metricId string, period string, value int) (Limit, error) {
	endpoint := fmt.Sprintf(limitAppPlanCreate, appPlanId, metricId)

	values := url.Values{}
	values.Add("application_plan_id", appPlanId)

	return c.limitCreate(endpoint, accessToken, metricId, period, value, values)
}

// CreateLimitEndUserPlan - Adds a limit to a metric of an end user plan
// All applications with the application plan (end_user_plan_id) will be constrained by this new limit on the metric (metric_id).
func (c *ThreeScaleClient) CreateLimitEndUserPlan(accessToken string, endUserPlanId string, metricId string, period string, value int) (Limit, error) {
	endpoint := fmt.Sprintf(limitEndUserPlanCreateList, endUserPlanId, metricId)

	values := url.Values{}
	values.Add("end_user_plan_id", endUserPlanId)

	return c.limitCreate(endpoint, accessToken, metricId, period, value, values)
}

// UpdateLimitsPerPlan - Updates a limit on a metric of an end user plan
// Valid params keys and their purpose are as follows:
// "period" - Period of the limit
// "value"  - Value of the limit
func (c *ThreeScaleClient) UpdateLimitPerAppPlan(accessToken string, appPlanId string, metricId string, limitId string, p Params) (Limit, error) {
	endpoint := fmt.Sprintf(limitAppPlanUpdateDelete, appPlanId, metricId, limitId)
	return c.updateLimit(endpoint, accessToken, p)
}

// UpdateLimitsPerMetric - Updates a limit on a metric of an application plan
// Valid params keys and their purpose are as follows:
// "period" - Period of the limit
// "value"  - Value of the limit
func (c *ThreeScaleClient) UpdateLimitPerEndUserPlan(accessToken string, userPlanId string, metricId string, limitId string, p Params) (Limit, error) {
	endpoint := fmt.Sprintf(limitEndUserPlanUpdateDelete, userPlanId, metricId, limitId)
	return c.updateLimit(endpoint, accessToken, p)
}

// DeleteLimitPerAppPlan - Deletes a limit on a metric of an application plan
func (c *ThreeScaleClient) DeleteLimitPerAppPlan(accessToken string, appPlanId string, metricId string, limitId string) error {
	endpoint := fmt.Sprintf(limitAppPlanUpdateDelete, appPlanId, metricId, limitId)
	return c.deleteLimit(endpoint, accessToken)
}

// DeleteLimitPerEndUserPlan - Deletes a limit on a metric of an end user plan
func (c *ThreeScaleClient) DeleteLimitPerEndUserPlan(accessToken string, userPlanId string, metricId string, limitId string) error {
	endpoint := fmt.Sprintf(limitEndUserPlanUpdateDelete, userPlanId, metricId, limitId)
	return c.deleteLimit(endpoint, accessToken)
}

// ListLimitsPerAppPlan - Returns the list of all limits associated to an application plan.
func (c *ThreeScaleClient) ListLimitsPerAppPlan(accessToken string, appPlanId string) (LimitList, error) {
	endpoint := fmt.Sprintf(limitAppPlanList, appPlanId)
	return c.listLimits(endpoint, accessToken)
}

// ListLimitsPerEndUserPlan - Returns the list of all limits associated to an end user plan.
func (c *ThreeScaleClient) ListLimitsPerEndUserPlan(accessToken string, endUserPlanId string, metricId string) (LimitList, error) {
	endpoint := fmt.Sprintf(limitEndUserPlanCreateList, endUserPlanId, metricId)
	return c.listLimits(endpoint, accessToken)
}

// ListLimitsPerMetric - Returns the list of all limits associated to a metric of an application plan
func (c *ThreeScaleClient) ListLimitsPerMetric(accessToken string, appPlanId string, metricId string) (LimitList, error) {
	endpoint := fmt.Sprintf(limitAppPlanMetricList, appPlanId, metricId)
	return c.listLimits(endpoint, accessToken)
}

func (c *ThreeScaleClient) limitCreate(ep string, accessToken string, metricId string, period string, value int, values url.Values) (Limit, error) {
	var apiResp Limit

	values.Add("access_token", accessToken)
	values.Add("metric_id", metricId)
	values.Add("period", period)
	values.Add("value", strconv.Itoa(value))

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return apiResp, httpReqError
	}
	resp, err := c.httpClient.Do(req)

	if err != nil {
		return apiResp, genRespErr("create limit", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return apiResp, genRespErr("create limit", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("create limit", err.Error())
	}
	return apiResp, nil
}

func (c *ThreeScaleClient) updateLimit(ep string, accessToken string, p Params) (Limit, error) {
	var l Limit
	values := url.Values{}
	values.Add("access_token", accessToken)
	for k, v := range p {
		values.Add(k, v)
	}

	body := strings.NewReader(values.Encode())
	req, err := c.buildUpdateReq(ep, body)
	if err != nil {
		return l, httpReqError
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return l, genRespErr("update limit", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return l, genRespErr("update limit", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&l); err != nil {
		return l, genRespErr("update limit", err.Error())
	}
	return l, nil

}

func (c *ThreeScaleClient) deleteLimit(ep string, accessToken string) error {
	values := url.Values{}
	values.Add("access_token", accessToken)

	body := strings.NewReader(values.Encode())
	req, err := c.buildDeleteReq(ep, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return genRespErr("delete limit", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return genRespErr("delete limit", handleErrResp(resp))
	}
	return nil
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
	if err != nil {
		return ml, genRespErr("list limits per plan", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ml, genRespErr("list limits per plan", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&ml); err != nil {
		return ml, genRespErr("list limits per plan", err.Error())
	}
	return ml, nil
}
