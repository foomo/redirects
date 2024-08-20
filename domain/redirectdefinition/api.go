package redirectdefinition

import (
	"context"
	"errors"

	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	redirectnats "github.com/foomo/redirects/pkg/nats"
	redirectprovider "github.com/foomo/redirects/pkg/provider"
	"go.uber.org/zap"
)

// API for the domain
type (
	API struct {
		l                         *zap.Logger
		qry                       Queries
		cmd                       Commands
		getSiteIdentifierProvider redirectprovider.SiteIdentifierProviderFunc
		repo                      redirectrepository.RedirectsDefinitionRepository
	}
	Option func(api *API)
)

func NewAPI(
	l *zap.Logger,
	repo redirectrepository.RedirectsDefinitionRepository,
	updateSignal *redirectnats.UpdateSignal,
	opts ...Option,
) (*API, error) {
	inst := &API{
		l:    l,
		repo: repo,
	}
	if inst.l == nil {
		return nil, errors.New("missing logger")
	}
	inst.cmd = Commands{
		CreateRedirects: redirectcommand.CreateRedirectsHandlerComposed(
			redirectcommand.CreateRedirectsHandler(inst.repo),
			redirectcommand.CreateRedirectsConsolidateMiddleware(repo, false),
			redirectcommand.CreateRedirectsAutoCreateMiddleware(),
			redirectcommand.CreateRedirectsPublishMiddleware(updateSignal),
		),
		CreateRedirect: redirectcommand.CreateRedirectHandlerComposed(
			redirectcommand.CreateRedirectHandler(inst.repo),
			redirectcommand.CreateRedirectPublishMiddleware(updateSignal),
		),
		UpdateRedirect: redirectcommand.UpdateRedirectHandlerComposed(
			redirectcommand.UpdateRedirectHandler(inst.repo),
			redirectcommand.UpdateRedirectPublishMiddleware(updateSignal),
		),
		DeleteRedirect: redirectcommand.DeleteRedirectHandlerComposed(
			redirectcommand.DeleteRedirectHandler(inst.repo),
			redirectcommand.DeleteRedirectPublishMiddleware(updateSignal),
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

func (a *API) CreateRedirects(ctx context.Context, cmd redirectcommand.CreateRedirects) error {
	return a.cmd.CreateRedirects(ctx, a.l, cmd)
}

func (a *API) CreateRedirect(ctx context.Context, cmd redirectcommand.CreateRedirect) error {
	return a.cmd.CreateRedirect(ctx, a.l, cmd)
}

func (a *API) UpdateRedirect(ctx context.Context, cmd redirectcommand.UpdateRedirect) error {
	return a.cmd.UpdateRedirect(ctx, a.l, cmd)
}

func (a *API) DeleteRedirect(ctx context.Context, cmd redirectcommand.DeleteRedirect) error {
	return a.cmd.DeleteRedirect(ctx, a.l, cmd)
}

func (a *API) GetRedirects(ctx context.Context) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
	return a.qry.GetRedirects(ctx, a.l)
}

func (a *API) Search(ctx context.Context, qry redirectquery.Search) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
	return a.qry.Search(ctx, a.l, qry)
}
