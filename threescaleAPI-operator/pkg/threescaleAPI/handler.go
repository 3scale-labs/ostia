package threescaleAPI

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/3scale/ostia/threescaleAPI-operator/pkg/apis/3scale/v1alpha1"
	"github.com/3scale/ostia/threescaleAPI-operator/pkg/threescale/system_client"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"net/http"
	"net/url"
	"strconv"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {

	case *v1alpha1.API:

		fmt.Println("Starting Run")

		swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromYAMLData([]byte(o.Spec.OpenAPIDefinition))
		if err != nil {
			panic(err)
		}

		//Extract Plans from Swagger
		desiredPlans, err := getPlansFromSwagger(swagger)
		if err != nil {
			panic(err)
		}

		// Extract Endpoints from Swagger
		desiredEndpoints, err := getEndpointsFromSwagger(swagger)
		if err != nil {
			panic(err)
		}

		c, err := createClientFromCrd(o)
		if err != nil {
			panic("error creating required client for 3scale system")
		}

		accessToken := o.Spec.The3ScaleConfig.AccessToken
		serviceName := o.Name

		service, err := ensureServiceExists(c, accessToken, serviceName)
		if err != nil {
			return fmt.Errorf("error synchronizing service state - %s", err.Error())

		}

		existingEndpoints, err := getEndpointsFrom3scaleSystem(c, accessToken, service)
		if err != nil {
			fmt.Printf("Couldn't get Endpoints from 3scale: %v\n", err)

		}

		existingPlans, err := getPlansFrom3scaleSystem(c, accessToken, service)

		if !compareEndpoints(desiredEndpoints, existingEndpoints) {
			fmt.Println("[!] Endpoints are not in sync.")
			err := reconcileEndpointsWith3scaleSystem(c, accessToken, service, existingEndpoints, desiredEndpoints)
			if err != nil {
				panic("something went wrong")
			}
		} else {
			fmt.Println("[=] Endpoints are in sync. Nothing to do.")
		}

		if !comparePlans(desiredPlans, existingPlans) {
			fmt.Println("[!] Plans are not in Sync")
			reconcilePlansAndLimits(c, service, accessToken, desiredPlans)

		} else {
			fmt.Println("[=] Plans are in sync. Nothing to do.")
		}

		fmt.Println("Run done.")

	}

	return nil
}

func createClientFromCrd(api *v1alpha1.API) (*client.ThreeScaleClient, error) {
	systemAdminPortalURL, err := url.Parse(api.Spec.The3ScaleConfig.AdminPortalURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing 3scale url from crd - %s", err.Error())
	}

	port := 0
	if systemAdminPortalURL.Port() == "" {
		switch scheme := systemAdminPortalURL.Scheme; scheme {
		case "http":
			port = 80
		case "https":
			port = 443
		}
	} else {
		port, err = strconv.Atoi(systemAdminPortalURL.Port())
		if err != nil {
			return nil, fmt.Errorf("admin portal URL invalid port - %s" + err.Error())
		}
	}

	adminPortal, err := client.NewAdminPortal(systemAdminPortalURL.Scheme, systemAdminPortalURL.Host, port)
	if err != nil {
		return nil, fmt.Errorf("invalid Admin Portal URL: %s", err.Error())
	}

	// TODO - This should be secure by default and overrideable for testing
	// TODO - Set some sensible default here to handle timeouts
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	insecureHttp := &http.Client{Transport: tr}

	c := client.NewThreeScale(adminPortal, insecureHttp)
	return c, nil
}
