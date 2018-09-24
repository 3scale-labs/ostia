package client

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// CreateLimit - Adds a limit to a metric of an application plan.
// All applications with the application plan (application_plan_id) will be constrained by this new limit on the metric (metric_id).
func (c *ThreeScaleClient) CreateLimit(accessToken string, appPlanId string, metricId string, period string, value int) (LimitResp, error) {
	var apiResp LimitResp
	if value < 1 {
		return apiResp, errors.New("value must be positive")
	}

	ep := genLimitEp(appPlanId, metricId)

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

func genLimitEp(appPlanId string, metricId string) string {
	return fmt.Sprintf(createLimitEndpoint, appPlanId, metricId)
}
