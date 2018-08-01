package apicast

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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
func DeploymentConfig(api *v1alpha1.API) *openshiftv1.DeploymentConfig {
	apicastLabels := labelsForAPIcast(api.Name)
	apicastConfig := createConfig(api)
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
	return deploymentConfig
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
func createConfig(api *v1alpha1.API) string {

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
	pc := []PolicyChain{upStreamPolicy, processRateLimitPolicies(api.Spec.RateLimits)}

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
	config, err := json.Marshal(apicastConfig)

	if err != nil {
		log.Errorf("Failed to serialize object %v", err)
	}

	return string(config)
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

func processRateLimitPolicies(limits []v1alpha1.RateLimit) PolicyChain {
	var fixedLimiters []FixedWindowRateLimiter
	var leakyLimiters []LeakyBucketRateLimiter
	var connLimiters []ConnectionRateLimiter

	for _, limit := range limits {
		rl, err := parseRateLimit(limit)
		if err != nil {
			log.Error(fmt.Sprintf("error processing rate limit rule for %s - %s", limit.Name, err.Error()))
			continue
		}

		switch rl.(type) {
		case FixedWindowRateLimiter:
			fwRl := rl.(FixedWindowRateLimiter)
			fixedLimiters = append(fixedLimiters, fwRl)
		case LeakyBucketRateLimiter:
			lbRl := rl.(LeakyBucketRateLimiter)
			leakyLimiters = append(leakyLimiters, lbRl)
		case ConnectionRateLimiter:
			crRl := rl.(ConnectionRateLimiter)
			connLimiters = append(connLimiters, crRl)
		default:
			log.Errorf("unknown rate limit type %T. ignoring", rl)
			continue
		}
	}

	var config PolicyChainConfiguration
	if len(fixedLimiters) > 0 {
		config.FixedWindowLimiters = &fixedLimiters
	}
	if len(leakyLimiters) > 0 {
		config.LeakyBucketLimiters = &leakyLimiters
	}
	if len(connLimiters) > 0 {
		config.ConnectionLimiters = &connLimiters
	}

	return PolicyChain{"apicast.policy.rate_limit", config}
}

func parseRateLimit(rl v1alpha1.RateLimit) (interface{}, error) {
	key := LimiterKey{rl.Name, "plain", "service"}
	if rl.Source != "" {
		switch rl.Source {
		//TODO - Add support for more sources here, jwt etc ..
		case "ip":
			key.Name = "{{remote_addr}}"
		default:
			log.Errorf("unknown source %s", rl.Source)
		}
		key.NameType = "liquid"
	}
	// limit is common to leaky bucket and fixed window algorithms
	if rl.Limit == "" {
		if err := verifyConnLimiterProps(rl); err != nil {
			return nil, errors.New("error in connection limiter syntax -" + err.Error())
		}
		return makeConnLimiterRateLimiter(*rl.Conn, *rl.Burst, *rl.Delay, key), nil
	}

	multiplier := 1
	parsedLimitVal := strings.Split(rl.Limit, "/")
	parsedReqs, err := strconv.Atoi(parsedLimitVal[0])
	if err != nil || parsedReqs < 0 {
		return nil, errors.New("limit value must be a non-negative integer")
	}

	if len(parsedLimitVal) == 2 {
		switch parsedLimitVal[1] {
		case "s":
			break
		case "m":
			multiplier = 60
		case "hr":
			multiplier = 60 * 60
		default:
			return nil, fmt.Errorf("unrecognised unit of time %s, defaulting to seconds", parsedLimitVal[1])
		}
	}

	if rl.Burst == nil {
		return makeFixedWindowRateLimiter(parsedReqs, multiplier, key), nil
	}

	if ok := verifyMinimumValue(*rl.Burst); !ok {
		return nil, errors.New("burst must be non-negative")

	}

	if parsedReqs/multiplier < 0 {
		return nil, errors.New("specified request rate (number per second) threshold must be non-negative")
	}

	return makeLeakyBucketRateLimiter(*rl.Burst, parsedReqs/multiplier, key), nil
}

func verifyConnLimiterProps(cl v1alpha1.RateLimit) error {
	if cl.Burst == nil || verifyMinimumValue(*cl.Burst) {
		return errors.New("burst must be set for connection limiter and have a minimum value of 0")
	}
	if cl.Conn == nil || verifyMinimumValue(*cl.Conn) {
		return errors.New("conn must be set for connection limiter and have a minimum value of 0")
	}
	if cl.Delay == nil || verifyMinimumValue(*cl.Delay) {
		return errors.New("delay must be set for connection limiter and have a minimum value of 0")
	}
	return nil
}

func verifyMinimumValue(val int) bool {
	if val < 0 {
		return false
	}
	return true
}

func makeConnLimiterRateLimiter(conn int, burst int, delay int, key LimiterKey) ConnectionRateLimiter {
	return ConnectionRateLimiter{
		Conn:  conn,
		Burst: burst,
		Delay: delay,
		Key:   key,
	}
}

func makeFixedWindowRateLimiter(count int, window int, key LimiterKey) FixedWindowRateLimiter {
	return FixedWindowRateLimiter{
		Count:  count,
		Window: window,
		Key:    key,
	}
}

func makeLeakyBucketRateLimiter(burst int, rate int, key LimiterKey) LeakyBucketRateLimiter {
	return LeakyBucketRateLimiter{
		Burst: burst,
		Rate:  rate,
		Key:   key,
	}
}
