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
	Expose    bool       `json:"expose"` //TODO: Make expose readonly after creation
	Hostname  string     `json:"hostname"`
	Endpoints []Endpoint `json:"endpoints"`
}

// APIStatus Contains the Status of the API object
type APIStatus struct { //TODO: Make this struct not user editable
	Deployed bool `json:"deployed"`
}

// Endpoint is a struct used to define the different upstream services
type Endpoint struct {
	Name string `json:"name"` // Not really needed?
	Host string `json:"host"`
	Path string `json:"path"`
}
