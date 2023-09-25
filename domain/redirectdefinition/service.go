package redirectdefinition

import (
	"github.com/foomo/contentserver/content"
	keelmongo "github.com/foomo/keel/persistence/mongo"
	"go.uber.org/zap"
)

// Potentially add Nats to service (still not sure)
type Service struct {
	l *zap.Logger
}

func NewService(l *zap.Logger, p *keelmongo.Persistor) (*Service, error) {
	return &Service{
		l: l,
	}, nil
}

func (rs *Service) CreateRedirectsFromContentserverexport(old, new map[string]*content.RepoNode) error {
	// TODO: Implement
	return nil
}

func (rs *Service) Search(dimension, id, path string) ([]*RedirectDefinition, error) {
	// TODO: Implement
	return nil, nil
}

func (rs *Service) Create(def *RedirectDefinition) error {
	// TODO: Implement
	return nil
}

func (rs *Service) Delete(id string) error {
	// TODO: Implement
	return nil
}

func (rs *Service) Update(def *RedirectDefinition) error {
	// TODO: Implement
	return nil
}
