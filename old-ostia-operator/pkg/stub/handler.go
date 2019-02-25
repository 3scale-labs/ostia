package stub

import (
	"context"
	"os"

	"github.com/3scale/ostia/ostia-operator/pkg/apicast"
	"github.com/3scale/ostia/ostia-operator/pkg/apis/ostia/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	log "github.com/sirupsen/logrus"
)

// NewHandler returns a Handler
func NewHandler() sdk.Handler {
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

// Handle takes care of handling the events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.API:

		// Let the owner reference take care of cleaning everything
		if event.Deleted {
			log.Infof("[%s] Delete event for API: %s", o.Namespace, o.Name)
			return nil
		}

		err := apicast.Reconcile(o)

		if err != nil {
			log.Errorf("Reconcile error %v", err)
		}
	}
	return nil
}
