package api

import (
	"context"
	"testing"

	ostiav1alpha1 "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func testAPIControllerDeploymentCreate(t *testing.T) {
	var (
		name            = "hello"
		namespace       = "test"
		replicas  int32 = 3
	)

	// A Memcached object with metadata and spec.
	api := &ostiav1alpha1.API{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: ostiav1alpha1.APISpec{
			Expose:   true,
			Hostname: "example.com",
			Endpoints: []ostiav1alpha1.Endpoint{
				{Name: "hello", Host: "https://example.com", Path: "/hello"},
			},
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{
		api,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(ostiav1alpha1.SchemeGroupVersion, api)
	routev1.Install(s)

	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	// Create a ReconcileMemcached object with the scheme and fake client.
	r := &ReconcileAPI{client: cl, scheme: s}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	apicast := types.NamespacedName{
		Name:      "apicast-" + name,
		Namespace: namespace,
	}

	// Check if deployment has been created and has the correct size.
	dep := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), apicast, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
	dsize := *dep.Spec.Replicas
	if dsize != replicas {
		t.Errorf("dep size (%d) is not the expected size (%d)", dsize, replicas)
	}
}
