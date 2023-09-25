package redirection

import (
	"github.com/foomo/contentserver/content"
	keelmongo "github.com/foomo/keel/persistence/mongo"
	"go.uber.org/zap"
)

// Potentially add Nats to service (still not sure)
type Redirection struct {
	l     *zap.Logger
	store RedirectsStore
}

func NewRedirectionService(l *zap.Logger, p *keelmongo.Persistor) (*Redirection, error) {
	store, err := NewRedirectsStore(l, p)
	if err != nil {
		return nil, err
	}

	return &Redirection{
		l:     l,
		store: *store,
	}, nil
}

func (rs *Redirection) CreateRedirectsFromContentserverexport(old, new map[string]*content.RepoNode) error {
	// TODO: Implement
	return nil
}

func (rs *Redirection) Search(dimension, id, path string) ([]*RedirectDefinition, error) {
	// TODO: Implement
	return nil, nil
}

func (rs *Redirection) Create(def *RedirectDefinition) error {
	// TODO: Implement
	return nil
}

func (rs *Redirection) Delete(id string) error {
	// TODO: Implement
	return nil
}

func (rs *Redirection) Update(def *RedirectDefinition) error {
	// TODO: Implement
	return nil
}
