package redirectdefinition

import (
	"context"
	"errors"
	"time"

	commandx "github.com/foomo/redirects/v2/domain/redirectdefinition/command"
	queryx "github.com/foomo/redirects/v2/domain/redirectdefinition/query"
	repositoryx "github.com/foomo/redirects/v2/domain/redirectdefinition/repository"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	natsx "github.com/foomo/redirects/v2/pkg/nats"
	providerx "github.com/foomo/redirects/v2/pkg/provider"
	"go.uber.org/zap"
)

// API for the domain
type (
	API struct {
		l                                         *zap.Logger
		qry                                       Queries
		cmd                                       Commands
		repo                                      repositoryx.RedirectsDefinitionRepository
		getSiteIdentifierProvider                 providerx.SiteIdentifierProviderFunc
		restrictedSourcesProvider                 providerx.RestrictedSourcesProviderFunc
		userProvider                              providerx.UserProviderFunc
		isAutomaticRedirectInitiallyStaleProvider providerx.IsAutomaticRedirectInitiallyStaleProviderFunc
	}
	Option func(api *API)
)

func NewAPI(
	l *zap.Logger,
	repo repositoryx.RedirectsDefinitionRepository,
	updateSignal *natsx.UpdateSignal,
	opts ...Option,
) (*API, error) {
	inst := &API{
		l:                         l,
		repo:                      repo,
		restrictedSourcesProvider: defaultRestrictedSourcesProvider,
		userProvider:              defaultUserProvider,
		isAutomaticRedirectInitiallyStaleProvider: defaultIsAutomaticRedirectInitiallyStaleProvider,
	}
	if inst.l == nil {
		return nil, errors.New("missing logger")
	}

	for _, opt := range opts {
		opt(inst)
	}

	inst.cmd = Commands{
		CreateRedirects: commandx.CreateRedirectsHandlerComposed(
			commandx.CreateRedirectsHandler(inst.repo),
			commandx.CreateRedirectsConsolidateMiddleware(repo, false),
			commandx.CreateRedirectsAutoCreateMiddleware(inst.isAutomaticRedirectInitiallyStaleProvider()),
			commandx.CreateRedirectsPublishMiddleware(updateSignal, repo),
		),
		CreateRedirect: commandx.CreateRedirectHandlerComposed(
			commandx.CreateRedirectHandler(inst.repo),
			commandx.ValidateRedirectMiddleware(inst.restrictedSourcesProvider, inst.repo),
			commandx.CreateRedirectPublishMiddleware(updateSignal, repo),
		),
		UpdateRedirect: commandx.UpdateRedirectHandlerComposed(
			commandx.UpdateRedirectHandler(inst.repo),
			commandx.ValidateUpdateRedirectMiddleware(inst.restrictedSourcesProvider, inst.repo),
			commandx.UpdateRedirectPublishMiddleware(updateSignal, repo),
		),
		DeleteRedirect: commandx.DeleteRedirectHandlerComposed(
			commandx.DeleteRedirectHandler(inst.repo),
			commandx.DeleteRedirectPublishMiddleware(updateSignal, repo),
		),
		UpdateRedirectsState: commandx.UpdateRedirectsStateHandlerComposed(
			commandx.UpdateRedirectsStateHandler(inst.repo),
			commandx.UpdateRedirectsStatePublishMiddleware(updateSignal, repo),
		),
	}
	inst.qry = Queries{
		GetRedirects: queryx.GetRedirectsHandlerComposed(
			queryx.GetRedirectsHandler(inst.repo),
		),
		Search: queryx.SearchHandlerComposed(
			queryx.SearchHandler(inst.repo),
		),
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (a *API) CreateRedirects(ctx context.Context, cmd commandx.CreateRedirects) error {
	return a.cmd.CreateRedirects(ctx, a.l, cmd)
}

func (a *API) CreateRedirect(ctx context.Context, cmd commandx.CreateRedirect) error {
	return a.cmd.CreateRedirect(ctx, a.l, cmd)
}

func (a *API) UpdateRedirect(ctx context.Context, cmd commandx.UpdateRedirect) error {
	return a.cmd.UpdateRedirect(ctx, a.l, cmd)
}

func (a *API) UpdateRedirectsState(ctx context.Context, cmd commandx.UpdateRedirectsState) error {
	return a.cmd.UpdateRedirectsState(ctx, a.l, cmd)
}

func (a *API) DeleteRedirect(ctx context.Context, cmd commandx.DeleteRedirect) error {
	return a.cmd.DeleteRedirect(ctx, a.l, cmd)
}

func (a *API) GetRedirects(ctx context.Context) (map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition, error) {
	return a.qry.GetRedirects(ctx, a.l)
}

func (a *API) Search(ctx context.Context, qry queryx.Search) (*storex.PaginatedResult, error) {
	return a.qry.Search(ctx, a.l, qry)
}

func (a *API) setLastUpdatedBy(ctx context.Context, definition *storex.RedirectDefinition) {
	if definition != nil {
		username := a.userProvider(ctx)
		if username == "" {
			username = "unknown"
		}

		definition.LastUpdatedBy = username
		definition.Updated = storex.NewDateTime(time.Now())
	}
}
