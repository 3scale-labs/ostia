package api

import (
	"context"
	"github.com/3scale/ostia/ostia-operator/pkg/apicast"
	ostiav2alpha1 "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v2alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_api")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new API Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

type NoAPITrigger func(handler.MapObject) []reconcile.Request

func (r NoAPITrigger) Map(o handler.MapObject) []reconcile.Request {
	return r(o)
}

var NoAPITriggerFunc NoAPITrigger = func(o handler.MapObject) []reconcile.Request {
	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{
			Namespace: o.Meta.GetNamespace(),
			Name:      "_NoAPI",
		}},
	}
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAPI{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("api-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to primary resource API
	err = c.Watch(&source.Kind{Type: &ostiav2alpha1.API{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &ostiav2alpha1.Operation{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: NoAPITriggerFunc})

	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &ostiav2alpha1.Server{}}, &handler.EnqueueRequestsFromMapFunc{ToRequests: NoAPITriggerFunc})

	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ostiav2alpha1.API{},
	})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileAPI{}

// ReconcileAPI reconciles a API object
type ReconcileAPI struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a API object and makes changes based on the state read
// and what is in the API.Spec
func (r *ReconcileAPI) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling API")

	// If the request comes from a watched event that's not an API object,
	// we should check all the API objects in that namespace to be sure there
	// are no changes required. Kind of hack to avoid the complexity and
	// traversing multiple OwnerReferences...
	if request.Name == "_NoAPI" {
		opts := client.ListOptions{}
		opts.InNamespace(request.Namespace)
		APIList := &ostiav2alpha1.APIList{}
		err := r.client.List(context.TODO(), &opts, APIList)
		if err != nil {
			reqLogger.Error(err, "error")
			return reconcile.Result{}, nil
		}

		for _, api := range APIList.Items {
			err := apicast.Reconcile(r.client, &api)
			if err != nil {
				reqLogger.Error(err, "error")
			}
		}

		return reconcile.Result{}, nil
	}

	api := &ostiav2alpha1.API{}

	err := r.client.Get(context.TODO(), request.NamespacedName, api)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	err = apicast.Reconcile(r.client, api)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
