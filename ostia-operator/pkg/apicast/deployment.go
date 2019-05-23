package apicast

import (
	"fmt"
	"os"

	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	extensions "k8s.io/api/extensions/v1beta1"

	"github.com/3scale/ostia/ostia-operator/pkg/apicast/standalone"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	b64 "encoding/base64"

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

func toDataURI(mime string, str []byte) string {
	return fmt.Sprintf("data:%s;base64,%s", mime, b64.URLEncoding.EncodeToString(str))
}

// DeploymentConfig returns an openshift deploymentConfig object for APIcast
func DeploymentConfig(api *ostia.API) (*appsv1.Deployment, error) {
	apicastLabels := labelsForAPIcast(api.Name)
	apicastConfig, err := standalone.CreateConfig(api)
	if err != nil {
		return nil, err
	}
	apicastName := apicastName(api)

	deploymentConfig := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apicastName,
			Namespace: api.Namespace,
			Labels:    apicastLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: nil,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"deployment": apicastName,
					"app":        "apicast"},
			},
			Strategy: appsv1.DeploymentStrategy{Type: "RollingUpdate"},
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
							Image:           apicastImage,
							ImagePullPolicy: v1.PullAlways,
							Name:            "apicast",
							Ports: []v1.ContainerPort{
								{ContainerPort: 8080, Name: "proxy", Protocol: "TCP"},
								{ContainerPort: 8090, Name: "management", Protocol: "TCP"},
							},
							Env: []v1.EnvVar{
								{Name: "APICAST_LOG_LEVEL", Value: "debug"},
								{Name: "APICAST_ENVIRONMENT", Value: "standalone"},
								{Name: "APICAST_CONFIGURATION", Value: toDataURI("application/json", apicastConfig)},
							},
							LivenessProbe:  newHTTPProbe("/status/live", 8090, 10, 5, 10),
							ReadinessProbe: newTCPProbe(8080, 15, 5, 30), // standalone management API does not support this
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
func Service(api *ostia.API) *v1.Service {

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
				"deployment": apicastName, "app": "apicast",
			},
		},
	}

	addOwnerRefToObject(service, asOwner(api))
	return service

}

func Ingress(api *ostia.API) *extensions.Ingress {
	APIcastLabels := labelsForAPIcast(api.Name)
	apicastName := apicastName(api)

	ingress := &extensions.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      apicastName,
			Namespace: api.Namespace,
			Labels:    APIcastLabels,
		},
		Spec: extensions.IngressSpec{
			Rules: []extensions.IngressRule{
				{
					Host: api.Spec.Hostname,
					IngressRuleValue: extensions.IngressRuleValue{
						HTTP: &extensions.HTTPIngressRuleValue{
							Paths: []extensions.HTTPIngressPath{
								{
									Backend: extensions.IngressBackend{
										ServiceName: apicastName,
										ServicePort: intstr.FromString("proxy"),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	addOwnerRefToObject(ingress, asOwner(api))

	return ingress
}

func addOwnerRefToObject(obj metav1.Object, ownerRef metav1.OwnerReference) {
	obj.SetOwnerReferences(append(obj.GetOwnerReferences(), ownerRef))
}

func asOwner(api *ostia.API) metav1.OwnerReference {
	trueVar := true
	return metav1.OwnerReference{
		APIVersion: api.APIVersion,
		Kind:       api.Kind,
		Name:       api.Name,
		UID:        api.UID,
		Controller: &trueVar,
	}
}

func apicastName(api *ostia.API) string {
	return "apicast-" + api.Name
}

func newHTTPProbe(path string, port int32, initDelay int32, timeout int32, period int32) *v1.Probe {
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

func newTCPProbe(port int32, initDelay int32, timeout int32, period int32) *v1.Probe {
	return &v1.Probe{
		Handler: v1.Handler{
			TCPSocket: &v1.TCPSocketAction{
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
