package e2e

import (
	goctx "context"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	appsv1 "k8s.io/api/apps/v1"
	extensions "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"testing"
	"time"

	"github.com/3scale/ostia/ostia-operator/pkg/apis"
	operator "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	dynclient "sigs.k8s.io/controller-runtime/pkg/client"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
	routerReload         = time.Second * 2
)

func TestAPI(t *testing.T) {
	apiList := &operator.APIList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "API",
			APIVersion: "ostia.3scale.net/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, apiList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("api-group", func(t *testing.T) {
		t.Run("Deploy", testDeploy)
		t.Run("Deploy2", testDeploy)
		t.Run("FixedRateLimit", testFixedRateLimit)
		t.Run("Reconcile", testReconcile)
	})
}

func deployAPISpec(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, spec operator.APISpec, name string) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}

	// create API custom resource
	API := &operator.API{
		TypeMeta: metav1.TypeMeta{
			Kind:       "API",
			APIVersion: "ostia.3scale.net/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: API.ObjectMeta.Name, Namespace: namespace}, API)

	API.Spec = spec

	if err != nil {
		// use TestCtx's create helper to create the object and add a cleanup function for the new object
		err = f.Client.Create(goctx.TODO(), API, nil)
		if err != nil {
			return err
		}
	} else {
		err = f.Client.Update(goctx.TODO(), API)
		if err != nil {
			return err
		}

		key, err := dynclient.ObjectKeyFromObject(API)

		if err != nil {
			return err
		}

		t.Logf("resource type %+v with namespace/name (%+v) updated\n", API.GetObjectKind().GroupVersionKind().Kind, key)
	}

	err = waitForAPI(t, f, API)

	if err != nil {
		return err
	}

	err = waitForDeployment(t, f, API)

	if err != nil {
		return err
	}

	time.Sleep(routerReload) // for OpenShift router to reload, sigh

	return nil
}

func waitForDeployment(t *testing.T, f *framework.Framework, API *operator.API) error {
	name := fmt.Sprintf("apicast-%s", API.ObjectMeta.Name)
	namespace := API.ObjectMeta.Namespace
	objectKey := types.NamespacedName{Name: name, Namespace: namespace}

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		deployment := &appsv1.Deployment{}
		err = f.Client.Get(goctx.TODO(), objectKey, deployment)

		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s Deployment\n", name)
				return false, nil
			}
			return false, err
		}

		status := deployment.Status

		if status.ObservedGeneration != deployment.Generation {
			return false, nil
		}

		t.Logf("Deployment generation %d (replicas %d, ready %d, updated %d)\n", deployment.Generation, status.Replicas, status.ReadyReplicas, status.UpdatedReplicas)

		if status.Replicas == status.ReadyReplicas && status.Replicas == status.UpdatedReplicas && status.Replicas == *deployment.Spec.Replicas {
			t.Logf("Deployment has correct number of replicas: %d\n", status.Replicas)
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	return nil

}

func waitForAPI(t *testing.T, f *framework.Framework, API *operator.API) error {
	name := API.ObjectMeta.Name
	namespace := API.ObjectMeta.Namespace
	objectKey := types.NamespacedName{Name: name, Namespace: namespace}

	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		api := &operator.API{}
		err = f.Client.Get(goctx.TODO(), objectKey, api)

		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s API\n", name)
				return false, nil
			}
			return false, err
		}

		if api.Status.ObservedGeneration != api.Generation {
			return false, nil
		}

		t.Logf("API generation %d (observed %d)\n", api.Generation, api.Status.ObservedGeneration)

		for _, condition := range api.Status.Conditions {
			switch condition.Type {
			case "Ready":
				return condition.Status == "true", nil
			}
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	t.Logf("API available (Generation %d)\n", API.Generation)

	return nil
}

func initCtx(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) {
	t.Parallel()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables

	// wait for memcached-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "ostia-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
}

func deployAPI(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, spec operator.APISpec, name string) {
	err := deployAPISpec(t, f, ctx, spec, name)

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(routerReload) // for OpenShift router to reload, sigh
}

func getNamespace(t *testing.T, ctx *framework.TestCtx) string {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(fmt.Errorf("could not get namespace: %v", err))
	}
	return namespace
}

func getHost(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, name string) string {
	namespace := getNamespace(t, ctx)

	ingress := extensions.Ingress{}
	err := f.Client.Get(goctx.TODO(), types.NamespacedName{Name: fmt.Sprintf("apicast-%s", name), Namespace: namespace}, &ingress)

	//time.Sleep(time.Second * 10)

	if err != nil {
		t.Fatal(err)
	}

	for _, rule := range ingress.Spec.Rules {
		return rule.Host
	}

	return ""
}

func httpGet(t *testing.T, protocol string, host string, path string) (*http.Response, error) {
	url := fmt.Sprintf("%s://%s%s", protocol, host, path)

	res, err := http.Get(url)

	if err == nil {
		t.Logf("GET %s (%d)", url, res.StatusCode)
		// t.Logf("Response %v", res)
	} else {
		// t.Logf("Error: %v", err)
	}

	return res, err
}

func makeHttpRequests(t *testing.T, host string, path string, count int, status int) {
	for i := 1; i <= count; i++ {
		res, err := httpGet(t, "http", host, path)

		if err != nil {
			t.Fatal(err)
		} else if res.StatusCode != status {
			t.Fatalf("Response %v should have status code %d", res, status)
		}
	}
}

func testFixedRateLimit(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	f := framework.Global
	initCtx(t, f, ctx)
	defer ctx.Cleanup()

	var spec = operator.APISpec{
		Expose:   true,
		Hostname: genHostname(t, ctx, "rate-limited"),
		Endpoints: []operator.Endpoint{
			{
				Name: "hello",
				Host: "https://echo-api.3scale.net",
				Path: "/hello",
			},
		},
		RateLimits: []operator.RateLimit{
			{Type: "FixedWindow", Name: "fixed", Limit: "10/m"},
		},
	}

	deployAPI(t, f, ctx, spec, "rate-limited")
	host := getHost(t, f, ctx, "rate-limited")

	makeHttpRequests(t, host, "/hello", 10, 200)
	makeHttpRequests(t, host, "/hello", 10, 429)
}

func testDeploy(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	f := framework.Global
	initCtx(t, f, ctx)
	defer ctx.Cleanup()

	var spec = operator.APISpec{
		Expose:   true,
		Hostname: genHostname(t, ctx, "test"),
		Endpoints: []operator.Endpoint{
			{
				Name: "hello",
				Host: "https://echo-api.3scale.net",
				Path: "/test",
			},
		},
	}

	deployAPI(t, f, ctx, spec, "test")
	host := getHost(t, f, ctx, "test")

	makeHttpRequests(t, host, "/test", 1, 200)
}

func genHostname(t *testing.T, ctx *framework.TestCtx, name string) string {
	ns := getNamespace(t, ctx)

	return fmt.Sprintf("%s.%s.lvh.me", name, ns)
}

func testReconcile(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	f := framework.Global
	initCtx(t, f, ctx)

	var spec = operator.APISpec{
		Expose:   true,
		Hostname: genHostname(t, ctx, "test"),
		Endpoints: []operator.Endpoint{
			{
				Name: "hello",
				Host: "https://echo-api.3scale.net",
				Path: "/test",
			},
		},
	}

	deployAPI(t, f, ctx, spec, "test")
	host := getHost(t, f, ctx, "test")

	makeHttpRequests(t, host, "/test", 1, 200)

	spec.Endpoints = []operator.Endpoint{
		{
			Name: "hello",
			Host: "https://echo-api.3scale.net",
			Path: "/hello",
		},
	}
	deployAPI(t, f, ctx, spec, "test")
	host = getHost(t, f, ctx, "test")
	makeHttpRequests(t, host, "/hello", 1, 200)
}
