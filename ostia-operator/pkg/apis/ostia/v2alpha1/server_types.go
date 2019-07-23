package v2alpha1

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Name       string             `json:"name,omitempty"`
	BasePath   string             `json:"basePath"`
	TargetPort intstr.IntOrString `json:"targetPort"`
	TLS        bool               `json:"tls"`
}

type ServerTSLContext struct {
}

func (s *Server) resolveServerService(client client.Client) (*url.URL, error) {

	if s.Spec.Service != nil {
		protocol := "http"
		port := int32(80)
		host := s.Spec.Service.Name
		basePath := s.Spec.Service.BasePath
		var err error

		// Set the proper protocol if TLS is enabled.
		if s.Spec.Service.TLS {
			protocol = "https"
		}

		// TargetPort supports either INT or String, we need to handle it:
		switch s.Spec.Service.TargetPort.Type {

		// If targetPort is a string, that means we need to resolve the objreference,
		// and look for the proper port in the destination service.
		case intstr.String:

			targetPort := s.Spec.Service.TargetPort.StrVal
			service := v1.Service{}
			serviceName := types.NamespacedName{
				Namespace: s.Namespace,
				Name:      s.Spec.Service.Name,
			}
			err := client.Get(context.TODO(), serviceName, &service)

			if err != nil {
				return &url.URL{}, fmt.Errorf("invalid target service")
			}

			port = -1
			for _, sp := range service.Spec.Ports {
				if sp.Name == targetPort {
					port = sp.Port
					break
				}
			}

			if port == -1 {
				return &url.URL{}, fmt.Errorf("invalid service target port")
			}

		// If targetPort is an INT, let's just return it.
		case intstr.Int:
			port = s.Spec.Service.TargetPort.IntVal
		}

		u, err := url.Parse(fmt.Sprintf("%s://%s:%d/%s", protocol, host, port, basePath))
		return u, err
	}
	return &url.URL{}, fmt.Errorf("no valid service")
}

func (s *Server) GetUpstream(client client.Client) (*url.URL, error) {

	// Spec.URL takes precedence. If this is not nil, lets construct and return this one.
	if s.Spec.URL != nil {
		u, err := url.Parse(*s.Spec.URL)
		return u, err
	}

	// Then get the target Service and see if the info is complete.
	if s.Spec.Service != nil {
		u, err := s.resolveServerService(client)
		return u, err
	}

	return &url.URL{}, fmt.Errorf("missing upstream")
}

func init() {
	SchemeBuilder.Register(&Server{}, &ServerList{})
}
