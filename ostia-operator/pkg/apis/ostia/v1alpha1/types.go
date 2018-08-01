package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIList is a list of API objects with extra Metadata
type APIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []API `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// API struct is used to define an API object
type API struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              APISpec   `json:"spec"`
	Status            APIStatus `json:"status,omitempty"`
}

// APISpec Contains the Spec of the API object
type APISpec struct {
	Expose     bool        `json:"expose"` //TODO: Make expose readonly after creation
	Hostname   string      `json:"hostname"`
	Endpoints  []Endpoint  `json:"endpoints"`
	RateLimits []RateLimit `json:"rate_limits,omitempty"`
}

// APIStatus Contains the Status of the API object
type APIStatus struct { //TODO: Make this struct not user editable
	Deployed bool `json:"deployed"`
}

// Endpoint is a struct used to define the different upstream services
type Endpoint struct {
	Name       string      `json:"name"` // Not really needed?
	Host       string      `json:"host"`
	Path       string      `json:"path"`
	RateLimits []RateLimit `json:"rate_limits,omitempty"`
}

// RateLimit is a struct used to define different types of rate limiting rules
type RateLimit struct {
	Burst  *int   `json:"burst"`
	Conn   *int   `json:"conn"`
	Delay  *int   `json:"delay"`
	Limit  string `json:"limit"`
	Name   string `json:"name"`   //TODO - This needs to reference and endpoint name currently but this relationship will reverse.
	Source string `json:"source"` // Source will allow user to limit based on jwt, source ip etc
}
