package redirectdefinition

import (
	"context"

	"github.com/foomo/contentserver/content"
	keelmongo "github.com/foomo/keel/persistence/mongo"
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
	service "github.com/foomo/redirects/domain/redirectdefinition/service"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// Potentially add Nats to service (still not sure)
type Service struct {
	l   *zap.Logger
	api *API
	ctx context.Context
}

func NewService(l *zap.Logger, p *keelmongo.Persistor, api *API, ctx context.Context) service.RedirectDefinitionService {
	return &Service{
		l:   l,
		api: api,
		ctx: ctx,
	}
}

func (rs *Service) CreateRedirectsFromContentserverexport(old, new map[string]*content.RepoNode) error {
	err := rs.api.CreateRedirects(context.Background(),
		redirectcommand.CreateRedirects{
			OldState: old,
			NewState: new,
		})
	if err != nil {
		return err
	}
	return nil
}

func (rs *Service) GetRedirects() (*redirectstore.RedirectDefinitions, error) {
	redirects, err := rs.api.GetRedirects(rs.ctx)
	if err != nil {
		return nil, err
	}
	return redirects, nil
}

func (rs *Service) Search(dimension, id, path string) (*redirectstore.RedirectDefinition, error) {
	definition, err := rs.api.Search(rs.ctx, redirectquery.Search{})
	if err != nil {
		return nil, err
	}
	return definition, nil
}

func (rs *Service) Create(def *redirectstore.RedirectDefinition) error {
	err := rs.api.CreateRedirect(rs.ctx,
		redirectcommand.CreateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return err
	}
	return nil
}

func (rs *Service) Delete(id string) error {
	err := rs.api.DeleteRedirect(rs.ctx,
		redirectcommand.DeleteRedirect{Source: redirectstore.RedirectSource(id)})
	if err != nil {
		return err
	}
	return nil
}

func (rs *Service) Update(def *redirectstore.RedirectDefinition) error {
	err := rs.api.UpdateRedirect(rs.ctx,
		redirectcommand.UpdateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return err
	}
	return nil
}
