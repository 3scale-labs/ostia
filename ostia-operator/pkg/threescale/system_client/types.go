package client

import (
	"encoding/xml"
	"net/http"
	"net/url"
)

// AdminPortal defines a 3scale adminPortal service
type AdminPortal struct {
	scheme  string
	host    string
	port    int
	baseUrl *url.URL
}

// ThreeScaleClient interacts with 3scale Service Management API
type ThreeScaleClient struct {
	adminPortal *AdminPortal
	httpClient  *http.Client
}

// ApplicationResp - API response for create limit endpoint
type ApplicationResp struct {
	Name                    xml.Name `xml:",any"`
	ID                      string   `xml:"id"`
	CreatedAt               string   `xml:"created_at"`
	UpdatedAt               string   `xml:"updated_at"`
	State                   string   `xml:"state"`
	UserAccountID           string   `xml:"user_account_id"`
	FirstTrafficAt          string   `xml:"first_traffic_at"`
	FirstDailyTrafficAt     string   `xml:"first_daily_traffic_at"`
	EndUserRequired         string   `xml:"end_user_required"`
	ServiceID               string   `xml:"service_id"`
	UserKey                 string   `xml:"user_key"`
	ProviderVerificationKey string   `xml:"provider_verification_key"`
	Plan                    struct {
		Text               string `xml:",chardata"`
		Custom             string `xml:"custom,attr"`
		Default            string `xml:"default,attr"`
		ID                 string `xml:"id"`
		Name               string `xml:"name"`
		Type               string `xml:"type"`
		State              string `xml:"state"`
		ServiceID          string `xml:"service_id"`
		EndUserRequired    string `xml:"end_user_required"`
		SetupFee           string `xml:"setup_fee"`
		CostPerMonth       string `xml:"cost_per_month"`
		TrialPeriodDays    string `xml:"trial_period_days"`
		CancellationPeriod string `xml:"cancellation_period"`
	} `xml:"plan"`
	AppName     string `xml:"name"`
	Description string `xml:"description"`
	ExtraFields string `xml:"extra_fields"`
	Error       string `xml:"error,omitempty"`
}

// Limit - Defines the object returned via the API for creation of a limit
type Limit struct {
	XMLName  xml.Name `xml:"limit"`
	ID       string   `xml:"id"`
	MetricID string   `xml:"metric_id"`
	PlanID   string   `xml:"plan_id"`
	Period   string   `xml:"period"`
	Value    string   `xml:"value"`
}

// LimitList - Holds a list of Limit
type LimitList struct {
	XMLName xml.Name `xml:"limits"`
	Limits  []Limit  `xml:"limit"`
}

// MappingRule - Defines the object returned via the API for creation of mapping rule
type MappingRule struct {
	XMLName    xml.Name `xml:"mapping_rule"`
	ID         string   `xml:"id,omitempty"`
	MetricID   string   `xml:"metric_id,omitempty"`
	Pattern    string   `xml:"pattern,omitempty"`
	HTTPMethod string   `xml:"http_method,omitempty"`
	Delta      string   `xml:"delta,omitempty"`
	CreatedAt  string   `xml:"created_at,omitempty"`
	UpdatedAt  string   `xml:"updated_at,omitempty"`
}

// MappingRuleList - Holds a list of MappingRule
type MappingRuleList struct {
	XMLName      xml.Name      `xml:"mapping_rules"`
	MappingRules []MappingRule `xml:"mapping_rule"`
}

// Metric - Defines the object returned via the API for creation of metric
type Metric struct {
	XMLName      xml.Name `xml:"metric"`
	ID           string   `xml:"id"`
	MetricName   string   `xml:"name"`
	SystemName   string   `xml:"system_name"`
	FriendlyName string   `xml:"friendly_name"`
	ServiceID    string   `xml:"service_id"`
	Description  string   `xml:"description"`
	Unit         string   `xml:"unit"`
}

// MetricList - Holds a list of Metric
type MetricList struct {
	XMLName xml.Name `xml:"metrics"`
	Metrics []Metric `xml:"metric"`
}

type AppPlansList struct {
	XMLName xml.Name `xml:"plans"`
	Text    string   `xml:",chardata"`
	Plan    []struct {
		Text               string `xml:",chardata"`
		Custom             string `xml:"custom,attr"`
		Default            string `xml:"default,attr"`
		ID                 string `xml:"id"`
		Name               string `xml:"name"`
		Type               string `xml:"type"`
		State              string `xml:"state"`
		ServiceID          string `xml:"service_id"`
		EndUserRequired    string `xml:"end_user_required"`
		SetupFee           string `xml:"setup_fee"`
		CostPerMonth       string `xml:"cost_per_month"`
		TrialPeriodDays    string `xml:"trial_period_days"`
		CancellationPeriod string `xml:"cancellation_period"`
	} `xml:"plan"`
}

// Plan - API response for create application plan endpoint
type Plan struct {
	Name               xml.Name `xml:",any"`
	Custom             string   `xml:"custom,attr"`
	Default            string   `xml:"default,attr"`
	ID                 string   `xml:"id"`
	PlanName           string   `xml:"name"`
	Type               string   `xml:"type"`
	State              string   `xml:"state"`
	ServiceID          string   `xml:"service_id"`
	EndUserRequired    string   `xml:"end_user_required"`
	SetupFee           string   `xml:"setup_fee"`
	CostPerMonth       string   `xml:"cost_per_month"`
	TrialPeriodDays    string   `xml:"trial_period_days"`
	CancellationPeriod string   `xml:"cancellation_period"`
	Error              string   `xml:"error,omitempty"`
}

type ServiceList struct {
	XMLName xml.Name `xml:"services"`
	Text    string   `xml:",chardata"`
	Service []struct {
		Text                        string `xml:",chardata"`
		ID                          string `xml:"id"`
		AccountID                   string `xml:"account_id"`
		Name                        string `xml:"name"`
		State                       string `xml:"state"`
		SystemName                  string `xml:"system_name"`
		BackendVersion              string `xml:"backend_version"`
		EndUserRegistrationRequired string `xml:"end_user_registration_required"`
		Metrics                     struct {
			Text   string   `xml:",chardata"`
			Metric []Metric `xml:"metric"`
			Method struct {
				Text         string `xml:",chardata"`
				ID           string `xml:"id"`
				Name         string `xml:"name"`
				SystemName   string `xml:"system_name"`
				FriendlyName string `xml:"friendly_name"`
				ServiceID    string `xml:"service_id"`
				Description  string `xml:"description"`
				MetricID     string `xml:"metric_id"`
			} `xml:"method"`
		} `xml:"metrics"`
	} `xml:"service"`
}

type ErrorResp struct {
	XMLName xml.Name `xml:"errors"`
	Text    string   `xml:",chardata"`
	Error   struct {
		Text string `xml:",chardata"`
	} `xml:"error"`
}

type Params map[string]string
