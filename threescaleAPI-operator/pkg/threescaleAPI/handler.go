package threescaleAPI

import (
	"context"
	"fmt"

	"github.com/3scale/ostia/threescaleAPI-operator/pkg/apis/3scale/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1alpha1.API:
		fmt.Println(o)
		return nil
	}
	return nil
}
