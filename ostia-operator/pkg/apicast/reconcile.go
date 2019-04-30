package apicast

import (
	"context"
	ostiav1alpha1 "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
)

//Reconcile takes care of the main apicast reconciliation loop
func Reconcile(client client.Client, request reconcile.Request) (err error) {
	// Fetch the API instance
	api := &ostiav1alpha1.API{}

	err = client.Get(context.TODO(), request.NamespacedName, api)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return nil
		}
		// Error reading the object - requeue the request.
		return err
	}
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Reconcile DeploymentConfig object
	err = reconcileDeployment(client, api)
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile DeploymentConfig")
	}

	// Reconcile Service object
	err = reconcileService(client, api)
	if err != nil {
		log.Error(err, "Failed to reconcile Service")
	}

	// Reconcile Route object
	if api.Spec.Expose {
		err = reconcileIngress(client, api)
		if err != nil {
			log.Error(err, "Failed to reconcile Route")
		}
	}

	return err
}

func namespacedName(meta v1.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      meta.GetName(),
		Namespace: meta.GetNamespace(),
	}
}

func reconcileDeployment(client client.Client, api *v1alpha1.API) (err error) {

	existingDc, err := Deployment(api)
	if err != nil {
		log.Error(err, "Failed to reconcile DeploymentConfig")
		return err
	}

	desiredDc, err := Deployment(api)
	if err != nil {
		return err
	}

	err = client.Get(context.TODO(), namespacedName(existingDc), existingDc)

	if err != nil {
		err = client.Create(context.TODO(), desiredDc)
		log.Info("Creating DeploymentConfig", "Error", err)
	} else {
		if !reflect.DeepEqual(existingDc.Spec, desiredDc.Spec) {
			existingDc.Spec = desiredDc.Spec
			err = client.Update(context.TODO(), existingDc)
			log.Info("Updating DeploymentConfig", "Error", err)
		}
	}

	return err
}

func reconcileService(client client.Client, api *v1alpha1.API) (err error) {

	existingSvc := Service(api)
	desiredSvc := Service(api)

	err = client.Get(context.TODO(), namespacedName(existingSvc), existingSvc)
	if err != nil {
		err = client.Create(context.TODO(), desiredSvc)
	} else {
		if !reflect.DeepEqual(existingSvc.Spec.Ports, desiredSvc.Spec.Ports) {
			existingSvc.Spec.Ports = desiredSvc.Spec.Ports
			err = client.Update(context.TODO(), existingSvc)
		}
	}
	return err

}

func reconcileIngress(client client.Client, api *v1alpha1.API) (err error) {

	existingIngress := Ingress(api)
	desiredIngress := Ingress(api)

	err = client.Get(context.TODO(), namespacedName(existingIngress), existingIngress)
	if err != nil {
		err = client.Create(context.TODO(), desiredIngress)
	} else {

		if !reflect.DeepEqual(existingIngress.Spec, desiredIngress.Spec) {
			existingIngress.Spec = desiredIngress.Spec
			err = client.Update(context.TODO(), existingIngress)
		}
	}

	return err

}
