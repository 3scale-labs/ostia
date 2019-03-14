package apis

import (
	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"

	appsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = registerOpenShiftAPIGroups(AddToSchemes)
}

func registerOpenShiftAPIGroups(builder runtime.SchemeBuilder) runtime.SchemeBuilder {
	return append(builder,
		appsv1.Install,
		imagev1.Install,
		routev1.Install,
	)
}
