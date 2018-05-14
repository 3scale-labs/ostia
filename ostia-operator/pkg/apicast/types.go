package apicast

const (
	ApicastImage   = "quay.io/3scale/apicast"
	ApicastVersion = "3.2-stable"
)

// Config is the configuration for APIcast
type Config struct {
	Services []Services `json:"services"`
}

// Services defines the services object
type Services struct {
	Proxy Proxy `json:"proxy"`
}

// PolicyChain contains a policy name and it's configuration
type PolicyChain struct {
	Name          string                   `json:"name"`
	Configuration PolicyChainConfiguration `json:"configuration"`
}

// PolicyChainConfiguration contains a group of PolicyChainRule
type PolicyChainConfiguration struct {
	Rules []PolicyChainRule `json:"rules"`
}

// Proxy defines the proxy struct for APIcast configuration
type Proxy struct {
	PolicyChain []PolicyChain `json:"policy_chain"`
	Hosts       []string      `json:"hosts"`
}

// TODO: Not all Chain Rules have this struct.

//PolicyChainRule Defines the content of a rule
type PolicyChainRule struct {
	Regex string `json:"regex"`
	URL   string `json:"url"`
}
