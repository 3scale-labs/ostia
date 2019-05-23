package apicast

import (
	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"testing"
)

func TestDeploymentConfig(t *testing.T) {
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
		},
	}
	var _, err = DeploymentConfig(api)

	if err != nil {
		println("ERROR: ", err)
		t.FailNow()
	}
}
