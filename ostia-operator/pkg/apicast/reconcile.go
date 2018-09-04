package apicast

import (
	"reflect"

	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
)

//Reconcile takes care of the main apicast reconciliation loop
func Reconcile(api *v1alpha1.API) (err error) {

	log.Infof("[%s] Got API Object: %s", api.Namespace, api.Name)

	// Reconcile DeploymentConfig object
	err = reconcileDeploymentConfig(api)
	if err != nil {
		log.Errorf("[%s] API: %s Failed to reconcile DeploymenConfig %v", api.Namespace, api.Name, err)
	}

	// Reconcile Service object
	err = reconcileService(api)
	if err != nil {
		log.Errorf("[%s] API: %s Failed to reconcile Service %v", api.Namespace, api.Name, err)
	}

	// Reconcile Route object
	if api.Spec.Expose {
		err = reconcileRoute(api)
		if err != nil {
			log.Errorf("[%s] API: %s Failed to reconcile Route %v", api.Namespace, api.Name, err)
		}
	}

	return err
}

func reconcileDeploymentConfig(api *v1alpha1.API) (err error) {

	existingDc, _ := DeploymentConfig(api)
	desiredDc, err := DeploymentConfig(api)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}

	err = sdk.Get(existingDc)
	if err != nil {
		err = sdk.Create(desiredDc)
	} else {
		if !reflect.DeepEqual(existingDc.Spec, desiredDc.Spec) {
			existingDc.Spec = desiredDc.Spec
			err = sdk.Update(existingDc)
		}
	}

	return err
}

func reconcileService(api *v1alpha1.API) (err error) {

	existingSvc := Service(api)
	desiredSvc := Service(api)

	err = sdk.Get(existingSvc)
	if err != nil {
		err = sdk.Create(desiredSvc)
	} else {
		if !reflect.DeepEqual(existingSvc.Spec.Ports, desiredSvc.Spec.Ports) {
			existingSvc.Spec.Ports = desiredSvc.Spec.Ports
			err = sdk.Update(existingSvc)
		}
	}
	return err

}

func reconcileRoute(api *v1alpha1.API) (err error) {

	existingRoute := Route(api)
	desiredRoute := Route(api)

	err = sdk.Get(existingRoute)
	if err != nil {
		err = sdk.Create(desiredRoute)
	} else {

		if !reflect.DeepEqual(existingRoute.Spec, desiredRoute.Spec) {
			existingRoute.Spec = desiredRoute.Spec
			err = sdk.Update(existingRoute)
		}
	}

	return err

}
