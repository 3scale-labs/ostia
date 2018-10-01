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

		systemAdminPortalURL, _ := url.Parse(o.Spec.The3ScaleConfig.AdminPortalURL)
		accessToken := o.Spec.The3ScaleConfig.AccessToken
		serviceName := o.Name
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
				fmt.Printf("Admin Portal URL invalid port: %v\n", err)
			}
		}
		adminPortal, err := client.NewAdminPortal(systemAdminPortalURL.Scheme, systemAdminPortalURL.Host, port)

		if err != nil {
			fmt.Printf("Invalid Admin Portal URL: %v\n", err)
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		insecureHttp := &http.Client{Transport: tr}

		c := client.NewThreeScale(adminPortal, insecureHttp)

		service, err := getServiceFromServiceSystemName(c, accessToken, serviceName)
		if err != nil {
			fmt.Printf("Service Name: %v not found: %v \n", o.Name, err)

		} else {

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
	}
	return nil
}
