package apicast

import (
	"encoding/json"
	"net/url"

	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	openshiftv1 "github.com/openshift/api/apps/v1"
	routev1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

//TODO: Define proper labels.
func labelsForAPIcast(name string) map[string]string {
	return map[string]string{"app": "apicast", "apiRef": name}
}

// DeploymentConfig returns an openshift deploymentConfig object for APIcast
func DeploymentConfig(api *v1alpha1.API) (*openshiftv1.DeploymentConfig, error) {
	apicastLabels := labelsForAPIcast(api.Name)
	apicastConfig, err := createConfig(api)
	if err != nil {
		return nil, err
	}
	apicastName := apicastName(api)

	deploymentConfig := &openshiftv1.DeploymentConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps.openshift.io/v1",
			Kind:       "DeploymentConfig",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apicastName,
			Namespace: api.Namespace,
			Labels:    apicastLabels,
		},
		Spec: openshiftv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: map[string]string{
				"deploymentconfig": apicastName,
			},
			Strategy: openshiftv1.DeploymentStrategy{
				Type: openshiftv1.DeploymentStrategyTypeRolling,
			},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deploymentconfig": apicastName,
						"app":              "apicast",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: apicastImage + ":" + apicastVersion,
							Name:  "apicast",
							Ports: []v1.ContainerPort{
								{ContainerPort: 8080, Name: "proxy", Protocol: "TCP"},
								{ContainerPort: 8090, Name: "management", Protocol: "TCP"},
							},
							Env: []v1.EnvVar{
								{Name: "APICAST_LOG_LEVEL", Value: "debug"},
								{Name: "THREESCALE_CONFIG_FILE", Value: "/tmp/load.json"},
								{Name: "APICAST_CONFIGURATION", Value: "data:application/json," + url.QueryEscape(apicastConfig)},
							},
							LivenessProbe:  newProbe("/status/live", 8090, 10, 5, 10),
							ReadinessProbe: newProbe("/status/ready", 8090, 15, 5, 30),
						},
					},
				},
			},
		},
	}
	addOwnerRefToObject(deploymentConfig, asOwner(api))
	return deploymentConfig, nil
}

// Service returns a k8s service object for APIcast
func Service(api *v1alpha1.API) *v1.Service {

	apicastLabels := labelsForAPIcast(api.Name)
	apicastName := apicastName(api)

	service := &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apicastName,
			Namespace: api.Namespace,
			Labels:    apicastLabels,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{Name: "proxy", Port: 8080, Protocol: "TCP", TargetPort: intstr.FromInt(8080)},
				{Name: "management", Port: 8090, Protocol: "TCP", TargetPort: intstr.FromInt(8090)},
			},
			Selector: map[string]string{
				"deploymentconfig": apicastName,
			},
		},
	}

	addOwnerRefToObject(service, asOwner(api))
	return service

}

// Route returns an openshift Route object for APIcast
func Route(api *v1alpha1.API) *routev1.Route {

	APIcastLabels := labelsForAPIcast(api.Name)
	apicastName := apicastName(api)

	route := &routev1.Route{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "route.openshift.io/v1",
			Kind:       "Route",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apicastName,
			Namespace: api.Namespace,
			Labels:    APIcastLabels,
		},
		Spec: routev1.RouteSpec{
			Host: api.Spec.Hostname,
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: apicastName,
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

//createConfig returns an APIcast Configuration Object
func createConfig(api *v1alpha1.API) (string, error) {
	var config string
	var apicastRules []PolicyChainRule
	upstreamServices := make([]string, len(api.Spec.Endpoints))

	for _, v := range api.Spec.Endpoints {
		rule := PolicyChainRule{
			Regex: v.Path,
			URL:   v.Host,
		}
		upstreamServices = append(upstreamServices, v.Name)
		apicastRules = append(apicastRules, rule)
	}

	var apicastHosts []string

	if api.Spec.Expose {
		apicastHosts = append(apicastHosts, api.Spec.Hostname)
	}

	apicastHosts = append(apicastHosts, apicastName(api))
	upStreamPolicy := PolicyChain{"apicast.policy.upstream", PolicyChainConfiguration{Rules: &apicastRules}}
	rateLimits, err := processRateLimitPolicies(api.Spec.RateLimits)
	if err != nil {
		return config, err
	}
	pc := []PolicyChain{upStreamPolicy, rateLimits}

	apicastConfig := &Config{
		Services: []Services{
			{
				Proxy: Proxy{
					Hosts:       apicastHosts,
					PolicyChain: pc,
				},
			},
		},
	}
	b, err := json.Marshal(apicastConfig)
	config = string(b)

	if err != nil {
		log.Errorf("Failed to serialize object %v", err)
	}

	return string(config), nil
}

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

func apicastName(api *v1alpha1.API) string {
	return "apicast-" + api.Name
}

func newProbe(path string, port int32, initDelay int32, timeout int32, period int32) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: path,
				Port: intstr.IntOrString{IntVal: port},
			},
		},
		InitialDelaySeconds: initDelay,
		TimeoutSeconds:      timeout,
		PeriodSeconds:       period,
	}
}
