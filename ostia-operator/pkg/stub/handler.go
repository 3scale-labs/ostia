package stub

import (
	"github.com/3scale/ostia/ostia-operator/pkg/apicast"
	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk/action"
	"github.com/operator-framework/operator-sdk/pkg/sdk/handler"
	"github.com/operator-framework/operator-sdk/pkg/sdk/query"
	"github.com/operator-framework/operator-sdk/pkg/sdk/types"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"os"
	"reflect"
)

// NewHandler returns a Handler
func NewHandler() handler.Handler {
	return &Handler{}
}

// Handler definition
type Handler struct {
}

func init() {
	// Set log level based on env var.
	loglevel := os.Getenv("LOG_LEVEL")
	switch loglevel {
	case "WARNING":
		log.SetLevel(log.WarnLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

func (h *Handler) Handle(ctx types.Context, event types.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.API:
		var err error

		api := o

		log.Infof("[%s] Got API Object: %s", api.Namespace, api.Name)

		// Let the owner reference take care of cleaning everything
		if event.Deleted {
			log.Infof("[%s] Delete event for API: %s", api.Namespace, api.Name)
			return nil
		}

		// Reconcile DeploymentConfig object
		existingDc := apicast.DeploymentConfig(api)
		desiredDc := apicast.DeploymentConfig(api)

		err = query.Get(existingDc)
		if err != nil {
			log.Infof("[%s] Failed to get: %v doesn't exists, trying to create", api.Namespace, err)

			err = action.Create(desiredDc)
			if err != nil && !apierrors.IsAlreadyExists(err) {
				log.Errorf("[%s] Failed to create DeployConfig: %v", api.Namespace, err)
			}

		} else {

			if !reflect.DeepEqual(existingDc.Spec, desiredDc.Spec) {
				existingDc.Spec = desiredDc.Spec
				err = action.Update(existingDc)
				if err != nil {
					log.Errorf("[%s] Failed to update: %v", api.Namespace, err)
				}
			}
		}

		// Reconcile Service object
		existingSvc := apicast.Service(api)
		desiredSvc := apicast.Service(api)

		err = query.Get(existingSvc)
		if err != nil {
			log.Infof("[%s] Failed to get: %v doesn't exists, trying to create", api.Namespace, err)

			err = action.Create(desiredSvc)
			if err != nil && !apierrors.IsAlreadyExists(err) {
				log.Errorf("[%s] Failed to create Service: %v", api.Namespace, err)
			}

		} else {

			if !reflect.DeepEqual(existingSvc.Spec.Ports, desiredSvc.Spec.Ports) {
				existingSvc.Spec.Ports = desiredSvc.Spec.Ports
				err = action.Update(existingSvc)
				if err != nil {
					log.Errorf("[%s] Failed to update: %v", api.Namespace, err)
				}
			}
		}

		// Reconcile Route object
		if api.Spec.Expose {
			existingRoute := apicast.Route(api)
			desiredRoute := apicast.Route(api)

			err = query.Get(existingRoute)
			if err != nil {
				log.Infof("[%s] Failed to get: %v doesn't exists, trying to create", api.Namespace, err)
				err = action.Create(desiredRoute)
				if err != nil && !apierrors.IsAlreadyExists(err) {
					log.Errorf("[%s] Failed to create Route: %v", api.Namespace, err)
				}
			} else {

				if !reflect.DeepEqual(existingRoute.Spec, desiredRoute.Spec) {
					existingRoute.Spec = desiredRoute.Spec
					err = action.Update(existingRoute)
					if err != nil {
						log.Errorf("[%s] Failed to update: %v", api.Namespace, err)
					}
				}
			}
		}

	}
	return nil
}
