package apicast

import (
	"context"
	ostia "github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v2alpha1"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("apicast")

func Reconcile(client client.Client, api *ostia.API) error {

	var err error

	reqLogger := log.WithValues("API.Namespace", api.Namespace, "API.Name", api.Name)

	if api.Generation != api.Status.ObservedGeneration {
		status := ostia.APIStatus{}
		status.Deployed = false
		status.ObservedGeneration = api.Generation
		status.Conditions = []ostia.APICondition{
			{Type: "Ready", Status: "false"},
		}

		if err := client.Status().Update(context.TODO(), api); err != nil {
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

	err = api.UpdateStatus(client)
	if err != nil {
		log.Error(err, "Failed to update API Status")
	}

	return nil
}

func reconcileDeploymentConfig(client client.Client, api *ostia.API) (err error) {
	existingDc, err := DeploymentConfig(client, api)

	if err != nil {
		log.Error(err, "Failed to reconcile Deployment")
		return err
	}

	desiredDc, err := DeploymentConfig(client, api)
	if err != nil {
		return err
	}

	err = client.Get(context.TODO(), namespacedName(existingDc), existingDc)

	if err != nil {
		err = client.Create(context.TODO(), desiredDc)
		log.Info("Creating Deployment", "Error", err)
	} else {

		if !compareDeployment(existingDc, desiredDc) {
			existingDc.Spec = desiredDc.Spec
			err = client.Update(context.TODO(), existingDc)
			log.Info("Updating Deployment", "Error", err)
		}
	}

	return err
}

func namespacedName(meta metav1.Object) types.NamespacedName {
	return types.NamespacedName{
		Name:      meta.GetName(),
		Namespace: meta.GetNamespace(),
	}
}

func compareDeployment(existing, desired *v1.Deployment) bool {

	if len(existing.Spec.Template.Spec.Containers) != len(desired.Spec.Template.Spec.Containers) {
		return false
	}

	if !reflect.DeepEqual(existing.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env) {
		return false
	}

	if existing.Spec.Template.Spec.Containers[0].Image != desired.Spec.Template.Spec.Containers[0].Image {
		return false
	}

	return true
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
