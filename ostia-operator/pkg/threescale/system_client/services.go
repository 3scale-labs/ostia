package client

import (
	"encoding/xml"
	"net/http"
	"net/url"
)

func (c *ThreeScaleClient) ListServices(accessToken string) (ServiceList, error) {
	var sl ServiceList

	ep := ServicesEndpoint

	req, err := c.buildGetReq(ep)
	if err != nil {
		return sl, httpReqError
	}

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()

	resp, err := c.httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return sl, genRespErr("List Services:", err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return sl, genRespErr("List Services:", handleErrResp(resp))
	}

	if err := xml.NewDecoder(resp.Body).Decode(&sl); err != nil {
		return sl, genRespErr("List Services:", err.Error())
	}
	return sl, nil
}
