package client

import (
	"encoding/xml"
	"fmt"
	"net/http"
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
func (c *ThreeScaleClient) CreateAppPlan(accessToken string, svcId string, name string, stateEvent string) (Plan, error) {
	var apiResp Plan
	ep := genAppPlanEp(svcId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("service_id", svcId)
	values.Add("name", name)
	values.Add("state_event", stateEvent)

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

func (c *ThreeScaleClient) ListAppPlanByServiceId(accessToken string, svcId string) (AppPlansList, error) {
		var appPlans AppPlansList
		ep := genAppPlansByService(svcId)

		req, err := c.buildGetReq(ep)
		if err != nil {
			return appPlans, httpReqError
		}

		values := url.Values{}
		values.Add("access_token", accessToken)
		values.Add("service_id", svcId)

		req.URL.RawQuery = values.Encode()
		resp, err := c.httpClient.Do(req)
		defer resp.Body.Close()

		if err != nil {
			return appPlans, genRespErr("List Application Plans By Service:", err.Error())
		}

		if resp.StatusCode != http.StatusOK {
			return appPlans, genRespErr("List Application Plans By Service:", handleErrResp(resp))
		}

		if err := xml.NewDecoder(resp.Body).Decode(&appPlans); err != nil {
			return appPlans, genRespErr("List Application Plans By Service:", err.Error())
		}
		return appPlans, nil
	}


func (c *ThreeScaleClient) ListAppPlan(accessToken string) (AppPlansList, error) {
	var appPlans AppPlansList
	ep := ListAppPlans

	req, err := c.buildGetReq(ep)
	if err != nil {
		return appPlans, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)

	req.URL.RawQuery = values.Encode()
	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return appPlans, genRespErr("List Application Plans By Service:", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return appPlans, genRespErr("List Application Plans By Service:", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&appPlans); err != nil {
		return appPlans, genRespErr("List Application Plans By Service:", err.Error())
	}
	return appPlans, nil
}

func genAppEp(accountId string) string {
	return fmt.Sprintf(createAppEndpoint, accountId)
}

func genAppPlanEp(svcId string) string {
	return fmt.Sprintf(createAppPlanEndpoint, svcId)
}

func genAppPlansByService(svcId string) string {
	return fmt.Sprintf(ListAppPlansByService, svcId)
}

