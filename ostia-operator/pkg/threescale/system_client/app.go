package client

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

// CreateAp - Create an application.
// The application object can be extended with Fields Definitions in the Admin Portal where you can add/remove fields
func (c *ThreeScaleClient) CreateApp(accessToken string, accountId string, planId string, name string, description string) (ApplicationResp, error) {
	var apiResp ApplicationResp
	ep := genAppEp(accountId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("account_id", accountId)
	values.Add("plan_id", planId)
	values.Add("name", name)
	values.Add("description", description)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return apiResp, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return apiResp, genRespErr("create application", err.Error())
	}

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("create application", err.Error())
	}
	return apiResp, nil
}

// CreateAppPlan - Creates an application plan.
func (c *ThreeScaleClient) CreateAppPlan(accessToken string, svcId string, name string) (PlanResp, error) {
	var apiResp PlanResp
	ep := genAppPlanEp(svcId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)
	values.Add("name", name)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(ep, body)
	if err != nil {
		return apiResp, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return apiResp, genRespErr("create application plan", err.Error())
	}

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("create application plan", err.Error())
	}
	return apiResp, nil
}

func genAppEp(accountId string) string {
	return fmt.Sprintf(createAppEndpoint, accountId)
}

func genAppPlanEp(svcId string) string {
	return fmt.Sprintf(createAppPlanEndpoint, svcId)
}
