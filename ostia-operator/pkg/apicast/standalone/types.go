package standalone

import ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"

type Configuration struct {
	Global    Global     `json:"global"`
	Server    Server     `json:"server"`
	Routes    []Route    `json:"routes"`
	Services  []Service  `json:"internal"`
	Upstreams []Upstream `json:"external"`
}

func NewConfiguration() Configuration {
	var config = Configuration{
		Global: Global{
			LogLevel:  "notice",
			ErrorLog:  "/dev/stderr",
			AccessLog: "/dev/stdout", // just stdout does not yet work in APIcast
		},
		Server: Server{
			Listen: []Listen{},
		},
		Routes:    []Route{},
		Services:  []Service{},
		Upstreams: []Upstream{},
	}

	return config
}

type Server struct {
	Listen []Listen `json:"listen"`
}

type Global struct {
	LogLevel  string `json:"log_level,omitempty"`
	ErrorLog  string `json:"error_log,omitempty"`
	AccessLog string `json:"access_log,omitempty"`
}

type Port = uint16
type Listen struct {
	Port          Port   `json:"port"`
	Name          string `json:"name"`
	ProxyProtocol bool   `json:"proxy_protocol,omitempty"`
	Protocol      string `json:"protocol,omitempty"` // http / spdy / http2
	TLS           *TLS   `json:"tls,omitempty"`
}

type TLS struct {
	Protocols      string   `json:"protocols"`
	Certificate    string   `json:"certificate"`
	CertificateKey string   `json:"certificate_key"`
	Ciphers        []string `json:"ciphers"`
}

type Route struct {
	Name        string      `json:"name"`
	Routes      []Route     `json:"routes,omitempty"`
	Match       Match       `json:"match"`
	Destination Destination `json:"destination"`
}

type Match struct {
	ServerPort string `json:"server_port,omitempty"`
	URIPath    string `json:"uri_path,omitempty"`
	HTTPMethod string `json:"http_method,omitempty"`
	HTTPHost   string `json:"http_host,omitempty"`
	Always     bool   `json:"always,omitempty"`
}

type Destination struct {
	Service      string `json:"service,omitempty"`
	Upstream     string `json:"upstream,omitempty"`
	HTTPResponse uint16 `json:"http_response,omitempty"`
}

type Policy struct {
	Name          string              `json:"policy"`
	Configuration PolicyConfiguration `json:"configuration,omitempty"`
}

type PolicyConfiguration = interface {
	//	MarshalJSON() ([]byte, error)
}

type Service struct {
	Name        string   `json:"name"`
	PolicyChain []Policy `json:"policy_chain"`
	Upstream    string   `json:"upstream,omitempty"`
}

type Upstream struct {
	Name         string `json:"name"`
	Server       string `json:"server"`
	LoadBalancer string `json:"load_balancer,omitempty"`
	Retries      uint8  `json:"retries,omitempty"`
}

func NewUpstream(server string) Upstream {
	var upstream = Upstream{
		Name:   server,
		Server: server,
	}

	return upstream
}

// PolicyChainConfiguration contains a group of PolicyChainRule
type RateLimitPolicyConfiguration struct {
	FixedWindowLimiters *[]FixedWindowRateLimiter `json:"fixed_window_limiters,omitempty"`
	LeakyBucketLimiters *[]LeakyBucketRateLimiter `json:"leaky_bucket_limiters,omitempty"`
	ConnectionLimiters  *[]ConnectionRateLimiter  `json:"connection_limiters,omitempty"`
}

var _ PolicyConfiguration = (*RateLimitPolicyConfiguration)(nil)

//FixedWindowRateLimiter defines a fixed window rate limiting rule
// Based on a fixed window of time (last X seconds).
// Can make up to Count requests per Window seconds.
type FixedWindowRateLimiter struct {
	Condition *ostia.Condition `json:"condition,omitempty"`
	Count     int              `json:"count"`
	Key       LimiterKey       `json:"key"`
	Window    int              `json:"window"`
}

//LeakyBucketRateLimiter defines a leaky bucket rate limiting rule
// Based on "leaky bucket" algorithm (average number of requests plus a maximum burst size)
// Can make up to Rate requests per second.
// It allows exceeding that number by Burst requests per second
// An artificial delay is introduced for those requests between rate and burst to avoid going over the limits.
type LeakyBucketRateLimiter struct {
	Burst     int              `json:"burst"`
	Condition *ostia.Condition `json:"condition,omitempty"`
	Key       LimiterKey       `json:"key"`
	Rate      int              `json:"rate"`
}

//ConnectionRateLimiter defines a connection rate, rate limiting rule
// Based on the concurrent number of connections.
// Conn is the max number of concurrent connections allowed.
// It allows exceeding that number by Burst connections per second.
// Delay is the number of seconds to delay the connections that exceed the limit.
type ConnectionRateLimiter struct {
	Burst     int              `json:"burst"`
	Condition *ostia.Condition `json:"condition,omitempty"`
	Conn      int              `json:"conn"`
	Delay     int              `json:"delay"`
	Key       LimiterKey       `json:"key"`
}

//LimiterKey defines a structure for rate limiting rules - name must be unique within scope
type LimiterKey struct {
	Name     string `json:"name"`
	NameType string `json:"name_type"`
	Scope    string `json:"scope"`
}

//LimiterCondition holds a set of conditions on which a limit will be applied
// assuming that the operations encapsulated by the condition holds true
type LimiterCondition struct {
	Operations ostia.Condition `json:"operations,omitempty"`
}
