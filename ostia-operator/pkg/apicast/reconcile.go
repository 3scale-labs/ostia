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

	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
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

	if api.Generation != api.Status.ObservedGeneration {
		status := ostia.APIStatus{}
		status.Deployed = false
		status.ObservedGeneration = api.Generation
		status.Conditions = []ostia.APICondition{
			{Type: "Ready", Status: "false"},
		}

		if err = client.Status().Update(context.TODO(), api); err != nil {
			return err
		}
	}

	// Reconcile DeploymentConfig object
	err = reconcileDeploymentConfig(client, api)
	if err != nil {
		reqLogger.Error(err, "Failed to reconcile Deployment")
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
			log.Error(err, "Failed to reconcile Ingress")
		}
	}

	err = updateStatus(client, api)

	if err != nil {
		log.Error(err, "Failed to update API Status")
	}

	return err
}

func updateStatus(client client.Client, api *ostia.API) (err error) {
	expectedStatus := *api.Status.DeepCopy()
	expectedStatus.Deployed = true
	expectedStatus.ObservedGeneration = api.Generation
	expectedStatus.Conditions = []ostia.APICondition{
		{Type: "Ready", Status: "true"},
	}

	if !reflect.DeepEqual(expectedStatus, api.Status) {
		log.Info("API Status does not match", "Expected", expectedStatus, "Actual", api.Status)

		api.Status = expectedStatus

		if err = client.Status().Update(context.TODO(), api); err != nil {
			return err
		}

		log.Info("Updated API Status", "APIStatus", expectedStatus)
	}

	return nil
}

func namespacedName(meta v1.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      meta.GetName(),
		Namespace: meta.GetNamespace(),
	}
}

func reconcileDeploymentConfig(client client.Client, api *ostia.API) (err error) {
	existingDc, err := DeploymentConfig(api)

	if err != nil {
		log.Error(err, "Failed to reconcile Deployment")
		return err
	}

	desiredDc, err := DeploymentConfig(api)
	if err != nil {
		return err
	}

	err = client.Get(context.TODO(), namespacedName(existingDc), existingDc)

	if err != nil {
		err = client.Create(context.TODO(), desiredDc)
		log.Info("Creating Deployment", "Error", err)
	} else {
		if !reflect.DeepEqual(existingDc.Spec, desiredDc.Spec) {
			existingDc.Spec = desiredDc.Spec
			err = client.Update(context.TODO(), existingDc)
			log.Info("Updating Deployment", "Error", err)
		}
	}

	return err
}

func reconcileService(client client.Client, api *ostia.API) (err error) {
	existingSvc := Service(api)
	desiredSvc := Service(api)

	err = client.Get(context.TODO(), namespacedName(existingSvc), existingSvc)
	if err != nil {
		err = client.Create(context.TODO(), desiredSvc)
		log.Info("Creating Service", "Error", err)
	} else {
		if !reflect.DeepEqual(existingSvc.Spec.Ports, desiredSvc.Spec.Ports) {
			existingSvc.Spec.Ports = desiredSvc.Spec.Ports
			err = client.Update(context.TODO(), existingSvc)
			log.Info("Updating Service", "Error", err)
		}
	}
	return err

}

func reconcileIngress(client client.Client, api *ostia.API) (err error) {
	existingIngress := Ingress(api)
	desiredIngress := Ingress(api)

	err = client.Get(context.TODO(), namespacedName(existingIngress), existingIngress)
	if err != nil {
		err = client.Create(context.TODO(), desiredIngress)
		log.Info("Creating Ingress", "Error", err)
	} else {

		if !reflect.DeepEqual(existingIngress.Spec, desiredIngress.Spec) {
			existingIngress.Spec = desiredIngress.Spec
			err = client.Update(context.TODO(), existingIngress)
			log.Info("Updating Ingress", "Error", err)
		}
	}

	return err
}
