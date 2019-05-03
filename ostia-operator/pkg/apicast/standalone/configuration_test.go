package standalone

import (
	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	var api = &ostia.API{
		Spec: ostia.APISpec{
			Expose:   true,
			Hostname: "example.com",
			Endpoints: []ostia.Endpoint{
				{
					Name: "hello",
					Host: "https://echo-api.3scale.net",
					Path: "/hello",
				},
			},
			RateLimits: []ostia.RateLimit{
				{
					Name:   "hello",
					Source: `{{remote_addr}}`,
					Limit:  "10/m",
					Type:   "FixedWindow",
				},
			},
		},
	}
	var standalone, err = CreateConfig(api)

	if err != nil {
		println("ERROR: ", err)
		t.FailNow()
	} else {
		println("SUCCESS: ", standalone)
	}
}
