package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type APIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []API `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type API struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              APISpec   `json:"spec"`
	Status            APIStatus `json:"status,omitempty"`
}

type The3ScaleConfig struct {
	AccessToken       string `json:"AccessToken"`
	AdminPortalURL    string `json:"AdminPortalURL"`
	IntegrationMethod string `json:"IntegrationMethod"`
}

type APISpec struct {
	The3ScaleConfig   The3ScaleConfig `json:"3scaleConfig"`
	Upstream          string          `json:"upstream"`
	OpenAPIDefinition string          `json:"OpenAPIDefinition"`
}
type APIStatus struct {
	// Fill me
}
