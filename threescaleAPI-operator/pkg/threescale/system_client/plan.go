package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	appPlanCreate         = "/admin/api/services/%s/application_plans.xml"
	appPlanUpdateDelete   = "/admin/api/services/%s/application_plans/%s.xml"
	appPlansList          = "/admin/api/application_plans.xml"
	appPlansByServiceList = "/admin/api/services/%s/application_plans.xml"
	appPlanSetDefault     = "/admin/api/services/%s/application_plans/%s/default.xml"
)

// CreateAppPlan - Creates an application plan.
func (c *ThreeScaleClient) CreateAppPlan(accessToken string, svcId string, name string, stateEvent string) (Plan, error) {
	var apiResp Plan
	endpoint := fmt.Sprintf(appPlanCreate, svcId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)
	values.Add("name", name)
	values.Add("state_event", stateEvent)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return apiResp, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiResp, genRespErr("create application plan", err.Error())
	}
	defer resp.Body.Close()

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("create application plan", err.Error())
	}
	return apiResp, nil
}

// UpdateAppPlan - Updates an application plan
func (c *ThreeScaleClient) UpdateAppPlan(accessToken string, svcId string, appPlanId string, name string, stateEvent string) (Plan, error) {
	endpoint := fmt.Sprintf(appPlanUpdateDelete, svcId, appPlanId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)
	values.Add("name", name)
	values.Add("state_event", stateEvent)

	return c.updatePlan(endpoint, accessToken, values)
}

// DeleteAppPlan - Deletes an application plan
func (c *ThreeScaleClient) DeleteAppPlan(accessToken string, svcId string, appPlanId string) error {
	endpoint := fmt.Sprintf(appPlanUpdateDelete, svcId, appPlanId)

	values := url.Values{}
	values.Add("access_token", accessToken)

	body := strings.NewReader(values.Encode())
	req, err := c.buildDeleteReq(endpoint, body)
	if err != nil {
		return httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return genRespErr("Delete App", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return genRespErr("Delete limit", handleErrResp(resp))
	}
	return nil
}

// ListAppPlanByServiceId - Lists all application plans, filtering on service id
func (c *ThreeScaleClient) ListAppPlanByServiceId(accessToken string, svcId string) (ApplicationPlansList, error) {
	var appPlans ApplicationPlansList
	endpoint := fmt.Sprintf(appPlansByServiceList, svcId)

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return appPlans, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)

	req.URL.RawQuery = values.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return appPlans, genRespErr("List Application Plans By Service:", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return appPlans, genRespErr("List Application Plans By Service:", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&appPlans); err != nil {
		return appPlans, genRespErr("List Application Plans By Service:", err.Error())
	}
	return appPlans, nil
}

// ListAppPlan - List all application plans
func (c *ThreeScaleClient) ListAppPlan(accessToken string) (ApplicationPlansList, error) {
	var appPlans ApplicationPlansList
	endpoint := appPlansList

	req, err := c.buildGetReq(endpoint)
	if err != nil {
		return appPlans, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)

	req.URL.RawQuery = values.Encode()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return appPlans, genRespErr("List Application Plans By Service:", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return appPlans, genRespErr("List Application Plans By Service:", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&appPlans); err != nil {
		return appPlans, genRespErr("List Application Plans By Service:", err.Error())
	}
	return appPlans, nil
}

// SetDefaultPlan - Makes the application plan the default one
func (c *ThreeScaleClient) SetDefaultPlan(accessToken string, svcId string, id string) (Plan, error) {
	endpoint := fmt.Sprintf(appPlanSetDefault, svcId, id)

	values := url.Values{}
	values.Add("access_token", accessToken)
	return c.updatePlan(endpoint, accessToken, values)
}

func (c *ThreeScaleClient) updatePlan(endpoint string, accessToken string, values url.Values) (Plan, error) {
	var apiResp Plan
	body := strings.NewReader(values.Encode())
	req, err := c.buildPutReq(endpoint, body)
	if err != nil {
		return apiResp, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiResp, genRespErr("Update application plan", err.Error())
	}
	defer resp.Body.Close()

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("Update application plan", err.Error())
	}
	return apiResp, nil
}
