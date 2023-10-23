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
	return a.cmd.CreateRedirects(ctx, a.l, cmd)
}

func (a *API) CreateRedirect(ctx context.Context, cmd redirectcommand.CreateRedirect) (err error) {
	return a.cmd.CreateRedirect(ctx, a.l, cmd)

}

func (a *API) UpdateRedirect(ctx context.Context, cmd redirectcommand.UpdateRedirect) (err error) {
	return a.cmd.UpdateRedirect(ctx, a.l, cmd)
}

func (a *API) DeleteRedirect(ctx context.Context, cmd redirectcommand.DeleteRedirect) (err error) {
	return a.cmd.DeleteRedirect(ctx, a.l, cmd)
}

func (a *API) GetRedirects(ctx context.Context) (redirects *redirectstore.RedirectDefinitions, err error) {
	return a.qry.GetRedirects(ctx, a.l)
}

func (a *API) Search(ctx context.Context, qry redirectquery.Search) (redirect *redirectstore.RedirectDefinitions, err error) {
	return a.qry.Search(ctx, a.l, qry)
}
