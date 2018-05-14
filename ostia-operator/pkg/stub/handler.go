package stub

import (
	"encoding/json"
	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	openshiftv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/operator-framework/operator-sdk/pkg/sdk/action"
	"github.com/operator-framework/operator-sdk/pkg/sdk/handler"
	"github.com/operator-framework/operator-sdk/pkg/sdk/query"
	"github.com/operator-framework/operator-sdk/pkg/sdk/types"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"
	"os"
	"reflect"
)

// NewHandler returns a Handler
func NewHandler() handler.Handler {
	return &Handler{}
}

// Handler definition
type Handler struct {
}

// TODO: Move apicast consts to APIcast pkg.

const (
	ApicastImage   = "quay.io/3scale/apicast"
	ApicastVersion = "3.2-stable"
)

// TODO: Move APIcast structs to their own pkg

// APIcastConfig is the configuration for APIcast
type APIcastConfig struct {
	Services []APIcastServices `json:"services"`
}

// APIcastServices defines the services object
type APIcastServices struct {
	Proxy APIcastProxy `json:"proxy"`
}

// APIcastPolicyChain contains a policy name and it's configuration
type APIcastPolicyChain struct {
	Name          string                          `json:"name"`
	Configuration APIcastPolicyChainConfiguration `json:"configuration"`
}

// APIcastPolicyChainConfiguration contains a group of APIcastPolicyChainRule
type APIcastPolicyChainConfiguration struct {
	Rules []APIcastPolicyChainRule `json:"rules"`
}

// APIcastProxy defines the proxy struct for APIcast configuration
type APIcastProxy struct {
	PolicyChain []APIcastPolicyChain `json:"policy_chain"`
	Hosts       []string             `json:"hosts"`
}

// TODO: Not all Chain Rules have this struct.
//APIcastPolicyChainRule Defines the content of a rule
type APIcastPolicyChainRule struct {
	Regex string `json:"regex"`
	URL   string `json:"url"`
}

func init() {
	// Set log level based on env var.
	loglevel := os.Getenv("LOG_LEVEL")
	switch loglevel {
	case "WARNING":
		log.SetLevel(log.WarnLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.API:
		var err error

		api := o

		log.Infof("[%s] Got API Object: %s", api.Namespace, api.Name)

		// Let the owner reference take care of cleaning everything
		if event.Deleted {
			log.Infof("[%s] Delete event for API: %s", api.Namespace, api.Name)
			return nil
		}

		// Reconcile DeploymentConfig object
		existingDc := deploymentConfigForAPIcast(api)
		desiredDc := deploymentConfigForAPIcast(api)

		err = query.Get(existingDc)
		if err != nil {
			log.Infof("[%s] Failed to get: %v doesn't exists, trying to create", api.Namespace, err)

			err = action.Create(desiredDc)
			if err != nil && !apierrors.IsAlreadyExists(err) {
				log.Errorf("[%s] Failed to create DeployConfig: %v", api.Namespace, err)
			}

		} else {

			if !reflect.DeepEqual(existingDc.Spec, desiredDc.Spec) {
				existingDc.Spec = desiredDc.Spec
				err = action.Update(existingDc)
				if err != nil {
					log.Errorf("[%s] Failed to update: %v", api.Namespace, err)
				}
			}
		}

		// Reconcile Service object
		existingSvc := serviceForAPIcast(api)
		desiredSvc := serviceForAPIcast(api)

		err = query.Get(existingSvc)
		if err != nil {
			log.Infof("[%s] Failed to get: %v doesn't exists, trying to create", api.Namespace, err)

			err = action.Create(desiredSvc)
			if err != nil && !apierrors.IsAlreadyExists(err) {
				log.Errorf("[%s] Failed to create Service: %v", api.Namespace, err)
			}

		} else {

			if !reflect.DeepEqual(existingSvc.Spec.Ports, desiredSvc.Spec.Ports) {
				existingSvc.Spec.Ports = desiredSvc.Spec.Ports
				err = action.Update(existingSvc)
				if err != nil {
					log.Errorf("[%s] Failed to update: %v", api.Namespace, err)
				}
			}
		}

		// Reconcile Route object
		if api.Spec.Expose {
			existingRoute := routeForAPIcast(api)
			desiredRoute := routeForAPIcast(api)

			err = query.Get(existingRoute)
			if err != nil {
				log.Infof("[%s] Failed to get: %v doesn't exists, trying to create", api.Namespace, err)
				err = action.Create(desiredRoute)
				if err != nil && !apierrors.IsAlreadyExists(err) {
					log.Errorf("[%s] Failed to create Route: %v", api.Namespace, err)
				}
			} else {

				if !reflect.DeepEqual(existingRoute.Spec, desiredRoute.Spec) {
					existingRoute.Spec = desiredRoute.Spec
					err = action.Update(existingRoute)
					if err != nil {
						log.Errorf("[%s] Failed to update: %v", api.Namespace, err)
					}
				}
			}
		}

	}
	return nil
}

// TODO: Create a utils pkg.
func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

func asOwner(api *v1alpha1.API) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.APIVersion,
		Kind:       api.Kind,
		Name:       api.Name,
		UID:        api.UID,
		Controller: &trueVar,
	}
}

//TODO: Define proper labels.
func labelsForAPIcast(name string) map[string]string {
	return map[string]string{"app": "apicast", "apiRef": name}
}

func deploymentConfigForAPIcast(api *v1alpha1.API) *openshiftv1.DeploymentConfig {

	APIcastLabels := labelsForAPIcast(api.Name)

	deploymentConfig := &openshiftv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.openshift.io/v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apicast-" + api.Name, //TODO: This is used everywhere and should be extracted to a func.
			Namespace: api.Namespace,
			Labels:    APIcastLabels,
		},
		Spec: openshiftv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentconfig": "apicast-" + api.Name,
			},
			Strategy: openshiftv1.DeploymentStrategy{
				Type: openshiftv1.DeploymentStrategyTypeRolling,
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": "apicast-" + api.Name,
						"app":              "apicast",
					},
				},
				Spec: v1.PodSpec{ //TODO: Add healthchecks
					Containers: []v1.Container{{
						Image: ApicastImage + ":" + ApicastVersion,
						Name:  "apicast",
						Ports: []v1.ContainerPort{
							{ContainerPort: 8080, Name: "proxy", Protocol: "TCP"},
							{ContainerPort: 8090, Name: "management", Protocol: "TCP"},
						},
						Env: []v1.EnvVar{
							{Name: "APICAST_LOG_LEVEL", Value: "debug"},
							{Name: "THREESCALE_CONFIG_FILE", Value: "/tmp/load.json"},
							{Name: "APICAST_CONFIGURATION", Value: "data:application/json," + url.QueryEscape(createAPIcastConfig(api))},
						},
					}},
				},
			},
		},
	}

	addOwnerRefToObject(deploymentConfig, asOwner(api))
	return deploymentConfig

}

func serviceForAPIcast(api *v1alpha1.API) *v1.Service {

	APIcastLabels := labelsForAPIcast(api.Name)

	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apicast-" + api.Name,
			Namespace: api.Namespace,
			Labels:    APIcastLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "proxy", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
				{Name: "management", Port: 8090, Protocol: "TCP", TargetPort: intstr.FromInt(8090)},
			},
			Selector: map[string]string{
				"deploymentconfig": "apicast-" + api.Name,
			},
		},
	}

	addOwnerRefToObject(service, asOwner(api))
	return service

}

func routeForAPIcast(api *v1alpha1.API) *routev1.Route {

	APIcastLabels := labelsForAPIcast(api.Name)

	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "route.openshift.io/v1",
			Kind:       "Route",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "apicast-" + api.Name,
			Namespace: api.Namespace,
			Labels:    APIcastLabels,
		},
		Spec: routev1.RouteSpec{
			Host: api.Spec.Hostname,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "apicast-" + api.Name,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("proxy"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyAllow,
			},
		},
	}

	addOwnerRefToObject(route, asOwner(api))
	return route

}

func createAPIcastConfig(api *v1alpha1.API) string {

	var apicastRules []APIcastPolicyChainRule

	for _, v := range api.Spec.Endpoints {
		rule := APIcastPolicyChainRule{
			Regex: v.Path,
			URL:   v.Host,
		}
		apicastRules = append(apicastRules, rule)
	}

	var apicastHosts []string

	if api.Spec.Expose {
		apicastHosts = append(apicastHosts, api.Spec.Hostname)
	}

	apicastHosts = append(apicastHosts, "apicast-"+api.Name)

	apicastConfig := &APIcastConfig{
		Services: []APIcastServices{
			{
				Proxy: APIcastProxy{
					Hosts: apicastHosts,
					PolicyChain: []APIcastPolicyChain{
						{
							Name: "apicast.policy.upstream",
							Configuration: APIcastPolicyChainConfiguration{
								Rules: apicastRules,
							},
						},
					},
				},
			},
		},
	}
	config, err := json.Marshal(apicastConfig)

	if err != nil {
		log.Errorf("Failed to serialize object %v", err)
	}

	return string(config)
}
