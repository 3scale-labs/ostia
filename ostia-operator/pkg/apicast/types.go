package apicast

const (
	apicastImage   = "quay.io/3scale/apicast"
	apicastVersion = "master"
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
	Rules               *[]PolicyChainRule        `json:"rules,omitempty"`
	FixedWindowLimiters *[]FixedWindowRateLimiter `json:"fixed_window_limiters,omitempty"`
	LeakyBucketLimiters *[]LeakyBucketRateLimiter `json:"leaky_bucket_limiters,omitempty"`
	ConnectionLimiters  *[]ConnectionRateLimiter  `json:"connection_limiters,omitempty"`
}

// Proxy defines the proxy struct for APIcast configuration
type Proxy struct {
	PolicyChain []PolicyChain `json:"policy_chain"`
	Hosts       []string      `json:"hosts"`
}

// TODO: Not all Chain Rules have this struct.

//PolicyChainRule Defines the content of a rule
type PolicyChainRule struct {
	Regex string `json:"regex,omitempty"`
	URL   string `json:"url,omitempty"`
}

//FixedWindowRateLimiter defines a fixed window rate limiting rule
// Based on a fixed window of time (last X seconds).
// Can make up to Count requests per Window seconds.
type FixedWindowRateLimiter struct {
	Count  int        `json:"count"`
	Key    LimiterKey `json:"key"`
	Window int        `json:"window"`
}

//LeakyBucketRateLimiter defines a leaky bucket rate limiting rule
// Based on "leaky bucket" algorithm (average number of requests plus a maximum burst size)
// Can make up to Rate requests per second.
// It allows exceeding that number by Burst requests per second
// An artificial delay is introduced for those requests between rate and burst to avoid going over the limits.
type LeakyBucketRateLimiter struct {
	Burst int        `json:"burst"`
	Key   LimiterKey `json:"key"`
	Rate  int        `json:"rate"`
}

//ConnectionRateLimiter defines a connection rate, rate limiting rule
// Based on the concurrent number of connections.
// Conn is the max number of concurrent connections allowed.
// It allows exceeding that number by Burst connections per second.
// Delay is the number of seconds to delay the connections that exceed the limit.
type ConnectionRateLimiter struct {
	Burst int        `json:"burst"`
	Conn  int        `json:"conn"`
	Delay int        `json:"delay"`
	Key   LimiterKey `json:"key"`
}

//LimiterKey defines a structure for rate limiting rules - name must be unique within scope
type LimiterKey struct {
	Name     string `json:"name"`
	NameType string `json:"name_type"`
	Scope    string `json:"scope"`
}
