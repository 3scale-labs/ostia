// NOTE: Boilerplate only.  Ignore this file.

// Package v2alpha1 contains API Schema definitions for the ostia v2alpha1 API group
// +k8s:deepcopy-gen=package,register
// +groupName=ostia.3scale.net
package v2alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "ostia.3scale.net", Version: "v2alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)
