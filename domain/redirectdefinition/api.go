package redirectdefinition

import (
	"context"
	"errors"

	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	"go.uber.org/zap"
)

// API for the domain
type (
	API struct {
		//qry  Queries
		cmd  Commands
		repo *redirectrepository.RedirectsDefinitionRepository
		l    *zap.Logger
		//meter                      *cmrccommonmetric.Meter
	}
	Option func(api *API)
)

func NewAPI(
	l *zap.Logger,
	repo *redirectrepository.RedirectsDefinitionRepository,
	opts ...Option,
) (*API, error) {

	inst := &API{
		l:    l,
		repo: repo,
		//meter:                      cmrccommonmetric.NewMeter(l, "checkout", telemetry.Meter()),
	}
	if inst.l == nil {
		return nil, errors.New("missing logger")
	}
	if inst.repo == nil {
		return nil, errors.New("missing cart repository")
	}
	inst.cmd = Commands{
		CreateRedirects: redirectcommand.CreateRedirectsHandlerComposed(
			redirectcommand.CreateRedirectsHandler(inst.repo),
		),
	}

	for _, opt := range opts {
		opt(inst)
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (a *API) CreateRedirects(ctx context.Context, cmd redirectcommand.CreateRedirects) (err error) {
	if err := a.cmd.CreateRedirects(ctx, a.l, cmd); err != nil {
		return err
	}
	return nil
}
