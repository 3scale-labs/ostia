package apis

import (
	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, ostia.SchemeBuilder.AddToScheme)
	AddToSchemes = registerOpenShiftAPIGroups(AddToSchemes)
}

func registerOpenShiftAPIGroups(builder runtime.SchemeBuilder) runtime.SchemeBuilder {
	return append(builder,
		appsv1.AddToScheme,
		imagev1.AddToScheme,
		routev1.AddToScheme,
	)
}
