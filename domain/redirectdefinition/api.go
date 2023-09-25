package redirectdefinition

import (
	"errors"

	"go.uber.org/zap"
)

// API for the domain
type (
	API struct {
		Queries  Queries
		Commands Commands
		//repo     *cmrccheckoutrepo.CheckoutRepository
		l *zap.Logger
		//meter                      *cmrccommonmetric.Meter
	}
	Option func(api *API)
)

func NewAPI(
	l *zap.Logger,
	opts ...Option,
) (*API, error) {

	inst := &API{
		l: l,
		//repo: checkoutRepo,
		//meter:                      cmrccommonmetric.NewMeter(l, "checkout", telemetry.Meter()),
	}
	if inst.l == nil {
		return nil, errors.New("missing logger")
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst, nil
}
