package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
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
	Plans             Plans           `json:"plans"`
}
type APIStatus struct {
	// Fill me
}

// TODO - If Plans are expected to be common across backends, this should be moved from here as appropriate
type Plans []Plan

func (p Plans) Sort() Plans {
	for _, plan := range p {
		sort.Slice(plan.Limits, func(i, j int) bool {
			if plan.Limits[i].Metric != plan.Limits[j].Metric {
				return plan.Limits[i].Metric < plan.Limits[j].Metric
			} else {
				return plan.Limits[i].Max < plan.Limits[j].Max
			}
		})
	}
	return p
}

type Plan struct {
	Default bool    `json:"default, omitempty"`
	Name    string  `json:"name"`
	Limits  []Limit `json:"limits"`
}

type Limit struct {
	Max    int64  `json:"max"`
	Metric string `json:"metric"`
	Period string `json:"period"`
}
