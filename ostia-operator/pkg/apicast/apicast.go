package apicast

import (
	"encoding/json"
	"fmt"
	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"
	"os"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("apicast")

const (
	defaultApicastImage   = "quay.io/3scale/apicast"
	defaultApicastVersion = "master"
)

var apicastImage = getProxyImageVersion()

//TODO: Define proper labels.
func labelsForAPIcast(name string) map[string]string {
	return map[string]string{"app": "apicast", "apiRef": name}
}

func Deployment(api *v1alpha1.API) (*v1beta2.Deployment, error) {
	apicastLabels := labelsForAPIcast(api.Name)
	apicastConfig, err := createConfig(api)
	if err != nil {
		return nil, err
	}
	apicastName := apicastName(api)
	numReplicas := int32(1)
	deployment := &v1beta2.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apicastName,
			Namespace: api.Namespace,
			Labels:    apicastLabels,
		},
		Spec: v1beta2.DeploymentSpec{
			Replicas: &numReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"deployment": apicastName,
					"app":        "apicast",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deployment": apicastName,
						"app":        "apicast",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Image: apicastImage,
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
			Strategy: v1beta2.DeploymentStrategy{
				Type: v1beta2.RollingUpdateDeploymentStrategyType,
			},
		},
	}

	addOwnerRefToObject(deployment, asOwner(api))
	return deployment, nil
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
				"deployment": apicastName,
			},
		},
	}

	addOwnerRefToObject(service, asOwner(api))
	return service

}

func Ingress(api *v1alpha1.API) *v1beta1.Ingress {
	APIcastLabels := labelsForAPIcast(api.Name)
	apicastName := apicastName(api)

	ingress := &v1beta1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "extensions/v1beta1",
			APIVersion: "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apicastName,
			Namespace: api.Namespace,
			Labels:    APIcastLabels,
		},
		Spec: v1beta1.IngressSpec{
			Backend: &v1beta1.IngressBackend{
				ServiceName: apicastName,
				ServicePort: intstr.FromString("proxy"),
			},
		},
	}

	addOwnerRefToObject(ingress, asOwner(api))
	return ingress
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
	pc := []PolicyChain{upStreamPolicy}

	if len(api.Spec.RateLimits) > 0 {
		rateLimits, err := processRateLimitPolicies(api.Spec.RateLimits)
		if err != nil {
			return config, err
		}
		pc = append(pc, rateLimits)
	}

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
		log.Error(err, "Failed to serialize object")
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

func getProxyImageVersion() string {
	image, imageSet := os.LookupEnv("APICAST_IMAGE")
	if !imageSet || image == "" {
		image = defaultApicastImage
	}

	tag, tagSet := os.LookupEnv("APICAST_VERSION")
	if !tagSet || tag == "" {
		tag = defaultApicastVersion
	}

	return fmt.Sprintf("%s:%s", image, tag)
}
