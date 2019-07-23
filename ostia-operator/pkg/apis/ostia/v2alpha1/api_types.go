package v2alpha1

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// APISpec defines the desired state of API
// +k8s:openapi-gen=true
type APISpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	Hostname          string               `json:"hostname"`
	Expose            bool                 `json:"expose"`
	ServerSelector    metav1.LabelSelector `json:"serverSelector"`
	OperationSelector metav1.LabelSelector `json:"operationSelector"`
}

type APIConditionType string

// APIStatus defines the observed state of API
// +k8s:openapi-gen=true

// APIStatus Contains the Status of the API object
type APIStatus struct {
	//TODO: Make this struct not user editable
	Deployed bool `json:"deployed"`

	// ObservedGeneration reflects the generation of the most recently observed ReplicaSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Represents the latest available observations of a replica set's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []APICondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

type APICondition struct {
	// Type of replica set condition.
	Type APIConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty"`
}


// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// API is the Schema for the apis API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type API struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   APISpec   `json:"spec,omitempty"`
	Status APIStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// APIList contains a list of API
type APIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []API `json:"items"`
}

func (api *API) UpdateStatus(client client.Client) (err error) {
	expectedStatus := *api.Status.DeepCopy()
	expectedStatus.Deployed = true
	expectedStatus.ObservedGeneration = api.Generation
	expectedStatus.Conditions = []APICondition{
		{Type: "Ready", Status: "true"},
	}

	if !reflect.DeepEqual(expectedStatus, api.Status) {
		api.Status = expectedStatus

		if err = client.Status().Update(context.TODO(), api); err != nil {
			return err
		}
	}

	return nil
}

func (api *API) ResolveSelectors(c client.Client) (ServerList, OperationList, error) {

	var err error
	serverList := ServerList{}
	operationList := OperationList{}

	opts := client.ListOptions{}
	opts.InNamespace(api.Namespace)
	opts.MatchingLabels(api.Spec.OperationSelector.MatchLabels)

	err = c.List(context.TODO(), &opts, &operationList)
	if err != nil {
		return ServerList{}, OperationList{}, err
	}
	opts.MatchingLabels(api.Spec.ServerSelector.MatchLabels)
	err = c.List(context.TODO(), &opts, &serverList)
	if err != nil {
		return ServerList{}, OperationList{}, err
	}

	return serverList, operationList, nil
}

func init() {
	SchemeBuilder.Register(&API{}, &APIList{})
}
