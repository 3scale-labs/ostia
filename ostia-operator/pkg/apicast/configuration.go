package apicast

import (
	"encoding/json"
	"github.com/3scale/ostia/ostia-operator/pkg/apicast/standalone"
	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v2alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// createConfig returns an APIcast Configuration Object
func createConfig(api *ostia.API, client client.Client) ([]byte, error) {
	var config = standalone.NewConfiguration()

	serverList, operationList, err := api.ResolveSelectors(client)
	if err != nil {
		return []byte{}, err
	}

	for _, server := range serverList.Items {

		upstream, err := server.GetUpstream(client)

		if err != nil {
			return []byte{}, err
		}

		s := standalone.Service{
			Name:        server.Name,
			PolicyChain: []standalone.Policy{},
			Upstream:    upstream.String(),
		}
		config.Services = append(config.Services, s)
	}

	for _, operation := range operationList.Items {

		r := standalone.Route{
			Name:   operation.Spec.ID,
			Routes: nil,
			Match: standalone.Match{
				ServerPort: "",
				URIPath:    operation.Spec.Path,
				HTTPMethod: operation.Spec.Method,
				HTTPHost:   api.Spec.Hostname,
				Always:     false,
			},
			Destination: standalone.Destination{
				Service:      operation.Spec.ServerRef,
				Upstream:     "",
				HTTPResponse: 0,
			},
		}

		config.Routes = append(config.Routes, r)
	}

	managementRoute := standalone.Route{
		Name:        "management",
		Match:       standalone.Match{ServerPort: "management"},
		Destination: standalone.Destination{Service: "management"},
	}

	config.Routes = append(config.Routes, managementRoute)

	config.Services = append(
		config.Services,
		standalone.Service{
			Name: "management", PolicyChain: []standalone.Policy{
				{Name: "apicast.policy.management"},
			},
		})

	if api.Spec.Expose {
		config.Server.Listen = []standalone.Listen{
			{Port: 8080, Name: "default", Protocol: "http"},
			{Port: 8090, Name: "management", Protocol: "http"},
		}
	}
	b, err := json.Marshal(config)

	if err != nil {
		log.Error(err, "Failed to serialize object")
		return b, err
	}

	return b, nil
}
