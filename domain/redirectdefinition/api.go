package redirectdefinition

import (
	"context"
	"errors"

	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// API for the domain
type (
	API struct {
		qry  Queries
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
		CreateRedirect: redirectcommand.CreateRedirectHandlerComposed(
			redirectcommand.CreateRedirectHandler(inst.repo),
		),
		UpdateRedirect: redirectcommand.UpdateRedirectHandlerComposed(
			redirectcommand.UpdateRedirectHandler(inst.repo),
		),
		DeleteRedirect: redirectcommand.DeleteRedirectHandlerComposed(
			redirectcommand.DeleteRedirectHandler(inst.repo),
		),
	}
	inst.qry = Queries{
		GetRedirects: redirectquery.GetRedirectsHandlerComposed(
			redirectquery.GetRedirectsHandler(inst.repo),
		),
		Search: redirectquery.SearchHandlerComposed(
			redirectquery.SearchHandler(inst.repo),
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

func (a *API) CreateRedirect(ctx context.Context, cmd redirectcommand.CreateRedirect) (err error) {
	if err := a.cmd.CreateRedirect(ctx, a.l, cmd); err != nil {
		return err
	}
	return nil
}

func (a *API) UpdareRedirect(ctx context.Context, cmd redirectcommand.UpdateRedirect) (err error) {
	if err := a.cmd.UpdateRedirect(ctx, a.l, cmd); err != nil {
		return err
	}
	return nil
}

func (a *API) DeleteRedirect(ctx context.Context, cmd redirectcommand.DeleteRedirect) (err error) {
	if err := a.cmd.DeleteRedirect(ctx, a.l, cmd); err != nil {
		return err
	}
	return nil
}

func (a *API) GetRedirects(ctx context.Context, qry redirectquery.GetRedirects) (redirects []*redirectstore.RedirectDefinition, err error) {
	if redirects, err = a.qry.GetRedirects(ctx, a.l, qry); err != nil {
		return nil, err
	}
	return redirects, err
}

func (a *API) Search(ctx context.Context, qry redirectquery.Search) (redirects []*redirectstore.RedirectDefinition, err error) {
	if redirects, err = a.qry.Search(ctx, a.l, qry); err != nil {
		return nil, err
	}
	return redirects, err
}
