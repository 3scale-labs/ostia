package standalone

import (
	"encoding/json"
	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("apicast-standalone")

func serviceValues(services map[string]Service) []Service {
	s := make([]Service, 0, len(services))

	for _, value := range services {
		s = append(s, value)
	}

	return s
}

//createConfig returns an APIcast Configuration Object
func CreateConfig(api *ostia.API) ([]byte, error) {
	var standalone = NewConfiguration()
	var rateLimit, err = processRateLimitPolicies(api.Spec.RateLimits)
	var routes = []Route{
		{
			Name:        "management",
			Match:       Match{ServerPort: "management"},
			Destination: Destination{Service: "management"}},
	}
	var services = make(map[string]Service)

	for _, v := range api.Spec.Endpoints {
		var service = Service{
			Name:        v.Host,
			PolicyChain: []Policy{rateLimit},
			Upstream:    v.Host,
		}
		routes = append(routes, Route{
			Name: v.Name,
			Match: Match{
				URIPath:    v.Path,
				ServerPort: "default",
			},
			Destination: Destination{Service: service.Name},
		})
		services[service.Name] = service
	}

	standalone.Routes = routes
	standalone.Services = append(
		serviceValues(services),
		Service{
			Name: "management", PolicyChain: []Policy{
				{Name: "apicast.policy.management"},
			},
		})

	if api.Spec.Expose {
		standalone.Server.Listen = []Listen{
			{Port: 8080, Name: "default", Protocol: "http"},
			{Port: 8090, Name: "management", Protocol: "http"},
		}
	}

	b, err := json.Marshal(standalone)

	if err != nil {
		log.Error(err, "Failed to serialize object")
		return b, err
	}

	return b, nil
}
