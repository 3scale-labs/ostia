package threescaleAPI

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/3scale/ostia/threescaleAPI-operator/pkg/apis/3scale/v1alpha1"
	"github.com/3scale/ostia/threescaleAPI-operator/pkg/threescale/system_client"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

func NewHandler(ns string) sdk.Handler {
	return &Handler{
		namespace: ns,
	}
}

type Handler struct {
	namespace string
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {

	case *v1alpha1.API:

		fmt.Printf("[i] Checking API: %s\n", o.Name)

		swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromYAMLData([]byte(o.Spec.OpenAPIDefinition))
		if err != nil {
			return fmt.Errorf("error loading swagger definition - API %s - error %v", o.Name, err)
		}

		//Extract Plans from Swagger
		desiredPlans := o.Spec.Plans

		// Extract Endpoints from Swagger
		desiredEndpoints, err := getEndpointsFromSwagger(swagger)
		if err != nil {
			return fmt.Errorf("error decoding desired endpoints from swagger definition - API %s - error %v", o.Name, err)
		}

		accessToken, portalUrl, err := extractClientCredentials(o, h.namespace)
		if err != nil {
			return fmt.Errorf("error extracting credentials required for 3scale system client - API %s - error %v", o.Name, err)
		}

		c, err := createSystemClient(portalUrl)
		if err != nil {
			return fmt.Errorf("error creating required client for 3scale system - API %s - error %v", o.Name, err)
		}

		serviceName := o.Name

		if event.Deleted {

			fmt.Printf("[-] Deleting service: %s\n", serviceName)

			service, err := getServiceFromServiceSystemName(c, accessToken, serviceName)
			if err != nil {
				if !strings.Contains(err.Error(), "not found") {
					return err
				}
			}

			err = c.DeleteService(accessToken, service.ID)
			if err != nil {
				return fmt.Errorf("Can't delete service %s, error: %s", serviceName, err.Error())
			}

		} else {

			service, err := ensureServiceExists(c, accessToken, serviceName)
			if err != nil {
				return fmt.Errorf("Error creating service: %s", err.Error())

			}

			// Calling proxy update here because if mapping rules have changed it needs to be called
			// The call is idempotent so if upstream is same as before then we receive a 200
			// This s less expensive than doing a proxy ready and a proxy update for the same effect
			p := client.NewParams()
			p.AddParam("api_backend", o.Spec.Upstream)
			_, err = c.UpdateProxy(accessToken, service.ID, p)
			if err != nil {
				fmt.Printf("Problem calling proxy update api. Desired changes may not be propogated. Error %v", err)
			}

			existingPlans, err := getPlansFrom3scaleSystem(c, accessToken, service)
			existingEndpoints, err := getEndpointsFrom3scaleSystem(c, accessToken, service)

			if !compareEndpoints(desiredEndpoints, existingEndpoints) {
				fmt.Println("[!] Endpoints are not in sync.")
				err := reconcileEndpointsWith3scaleSystem(c, accessToken, service, existingEndpoints, desiredEndpoints)
				if err != nil {
					panic("something went wrong")
				}

			} else {
				fmt.Println("[+] Endpoints are in sync. Nothing to do.")
			}

			if !comparePlans(desiredPlans, existingPlans) {
				fmt.Println("[!] Plans are not in Sync")
				reconcilePlansAndLimits(c, service, accessToken, desiredPlans)

			} else {
				fmt.Println("[+] Plans are in sync. Nothing to do.")
			}

			// We are not really checking the contents of the configuration, simply checking
			// for a mismatch in the version. Not sure about the value of improving this...

			productionProxy, _ := c.GetLatestProxyConfig(accessToken, service.ID, "production")
			sandboxProxy, _ := c.GetLatestProxyConfig(accessToken, service.ID, "sandbox")

			if productionProxy.ProxyConfig.Version != sandboxProxy.ProxyConfig.Version {
				fmt.Println("[!] Proxy Config is not in sync")
				_, err := c.PromoteProxyConfig(accessToken, service.ID, "sandbox", strconv.Itoa(sandboxProxy.ProxyConfig.Version), "production")
				if err != nil {
					return err
				}
			} else {
				fmt.Println("[+] Proxy Config is in sync. Nothing to do.")
			}

			fmt.Println("[i] Run done.")

		}

	}
	return nil

}

func extractClientCredentials(api *v1alpha1.API, namespace string) (accessToken, adminUrl string, err error) {
	if api.Spec.The3ScaleConfig.Credentials.Secret != nil {
		return extractCredentialsFromSecret(api, namespace)
	}

	accessToken = api.Spec.The3ScaleConfig.Credentials.AccessToken
	if accessToken == "" {
		return accessToken, adminUrl, fmt.Errorf("access token for 3scale system api must be set or provided via secret for api %s", api.Name)
	}

	adminUrl = api.Spec.The3ScaleConfig.Credentials.AdminPortalURL
	if adminUrl == "" {
		return accessToken, adminUrl, fmt.Errorf("admin portal url for 3scale system api must be set or provided via secret for api %s", api.Name)
	}

	return accessToken, adminUrl, nil
}

func extractCredentialsFromSecret(api *v1alpha1.API, namespace string) (accessToken, adminUrl string, err error) {
	secretName := api.Spec.The3ScaleConfig.Credentials.Secret.Name
	ns := api.Spec.The3ScaleConfig.Credentials.Secret.Namespace
	if ns == "" {
		ns = namespace
	}
	if secretName == "" {
		err = errors.New("secret name cannot be empty")
		return accessToken, adminUrl, err
	}

	secret := &coreV1.Secret{
		TypeMeta: metaV1.TypeMeta{
			Kind: "Secret",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      secretName,
			Namespace: ns,
		},
	}
	err = sdk.Get(secret)
	if err != nil {
		return accessToken, adminUrl, fmt.Errorf("error fetching secret %s - %s", secretName, err.Error())
	}

	if token, ok := secret.Data["access_token"]; ok {
		accessToken = string(token)
	} else {
		return accessToken, adminUrl, errors.New("'access_token key must be set in secret'")
	}

	if url, ok := secret.Data["admin_portal_url"]; ok {
		adminUrl = string(url)
	} else {
		return accessToken, adminUrl, errors.New("'admin_portal_url key must be set in secret'")
	}

	return accessToken, adminUrl, nil
}

func createSystemClient(portalUrl string) (*client.ThreeScaleClient, error) {
	systemAdminPortalURL, err := url.Parse(portalUrl)
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
