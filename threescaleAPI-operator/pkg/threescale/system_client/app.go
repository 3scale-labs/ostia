package client

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"strings"
)

const (
	appCreate = "/admin/api/accounts/%s/applications.xml"
)

// CreateAp - Create an application.
// The application object can be extended with Fields Definitions in the Admin Portal where you can add/remove fields
func (c *ThreeScaleClient) CreateApp(accessToken string, accountId string, planId string, name string, description string) (Application, error) {
	var apiResp Application
	endpoint := fmt.Sprintf(appCreate, accountId)

	values := url.Values{}
	values.Add("access_token", accessToken)
	values.Add("account_id", accountId)
	values.Add("plan_id", planId)
	values.Add("name", name)
	values.Add("description", description)

	body := strings.NewReader(values.Encode())
	req, err := c.buildPostReq(endpoint, body)
	if err != nil {
		return apiResp, httpReqError
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return apiResp, genRespErr("create application", err.Error())
	}
	defer resp.Body.Close()

	if err := xml.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return apiResp, genRespErr("create application", err.Error())
	}
	return apiResp, nil
}
