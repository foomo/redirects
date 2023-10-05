package redirectdefinition

import (
	"context"

	"github.com/foomo/contentserver/content"
	keelmongo "github.com/foomo/keel/persistence/mongo"
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// Potentially add Nats to service (still not sure)
type Service struct {
	l   *zap.Logger
	api *API
}

func NewService(l *zap.Logger, p *keelmongo.Persistor, api *API) (*Service, error) {
	return &Service{
		l:   l,
		api: api,
	}, nil
}

func (rs *Service) CreateRedirectsFromContentserverexport(old, new map[string]*content.RepoNode) error {
	// TODO: Implement
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

func (rs *Service) Search(dimension, id, path string) ([]*redirectstore.RedirectDefinition, error) {
	// TODO: Implement
	return nil, nil
}

func (rs *Service) Create(def *redirectstore.RedirectDefinition) error {
	// TODO: Implement
	return nil
}

func (rs *Service) Delete(id string) error {
	// TODO: Implement
	return nil
}

func (rs *Service) Update(def *redirectstore.RedirectDefinition) error {
	// TODO: Implement
	return nil
}
