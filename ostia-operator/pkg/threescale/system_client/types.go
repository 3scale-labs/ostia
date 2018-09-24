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

// LimitResp - API response for create limit endpoint
type LimitResp struct {
	Name     xml.Name `xml:",any"`
	ID       string   `xml:"id, omitempty"`
	MetricID string   `xml:"metric_id,omitempty"`
	PlanID   string   `xml:"plan_id,omitempty"`
	Period   string   `xml:"period,omitempty"`
	Value    string   `xml:"value,omitempty"`
	Error    string   `xml:"error,omitempty"`
}

// MappingRuleResp - API response for create mapping rule endpoint
type MappingRuleResp struct {
	Name       xml.Name `xml:",any"`
	ID         string   `xml:"id,omitempty"`
	MetricID   string   `xml:"metric_id,omitempty"`
	Pattern    string   `xml:"pattern,omitempty"`
	HTTPMethod string   `xml:"http_method,omitempty"`
	Delta      string   `xml:"delta,omitempty"`
	CreatedAt  string   `xml:"created_at,omitempty"`
	UpdatedAt  string   `xml:"updated_at,omitempty"`
	Error      string   `xml:"error,omitempty"`
}

// MetricResp - API response for create metric endpoint
type MetricResp struct {
	Name         xml.Name `xml:",any"`
	ID           string   `xml:"id"`
	MetricName   string   `xml:"name"`
	SystemName   string   `xml:"system_name"`
	FriendlyName string   `xml:"friendly_name"`
	ServiceID    string   `xml:"service_id"`
	Description  string   `xml:"description"`
	Unit         string   `xml:"unit"`
	Error        string   `xml:"error,omitempty"`
}

// PlanResp - API response for create application plan endpoint
type PlanResp struct {
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
