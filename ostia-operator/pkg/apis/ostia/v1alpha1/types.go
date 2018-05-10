package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type APIList struct {
	metav1.TypeMeta       `json:",inline"`
	metav1.ListMeta       `json:"metadata"`
	Items           []API `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type API struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec   APISpec    `json:"spec"`
	Status APIStatus  `json:"status,omitempty"`
}

type APISpec struct {
	Expose    bool       `json:"expose"` //TODO: Make expose readonly after creation
	Hostname  string     `json:"hostname"`
	Endpoints []Endpoint `json:"endpoints"`
}
type APIStatus struct {//TODO: Make this struct not user editable
	Deployed bool `json:"deployed"`
}

type Endpoint struct {
	Name string `json:"name"` // Not really needed?
	Host string `json:"host"`
	Path string `json:"path"`
}