package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServerSpec defines the desired state of Server
// +k8s:openapi-gen=true
type ServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
	RewriteHost string         `json:"rewriteHost,omitempty"`
	URL         *string        `json:"url,omitempty"`
	Default     *bool          `json:"default,omitempty"`
	Service     *ServerService `json:"service,omitempty"`
	//ServerTSLContext `json:"tlsContext,omitempty"`
}

// ServerStatus defines the observed state of Server
// +k8s:openapi-gen=true
type ServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Server is the Schema for the servers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerSpec   `json:"spec,omitempty"`
	Status ServerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServerList contains a list of Server
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Server `json:"items"`
}

// +k8s:openapi-gen=true

type ServerService struct {
	Selector   metav1.LabelSelector `json:"selector,omitempty"`
	BasePath   string               `json:"basePath"`
	TargetPort intstr.IntOrString   `json:"targetPort"`
	TLS        bool                 `json:"tls"`
}

type ServerTSLContext struct {
}

func init() {
	SchemeBuilder.Register(&Server{}, &ServerList{})
}
