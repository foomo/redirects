package redirectdefinition

import (
	"context"
	"net/http"

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
}

func NewService(l *zap.Logger, p *keelmongo.Persistor, api *API, ctx context.Context) service.RedirectDefinitionService {
	return &Service{
		l:   l,
		api: api,
	}
}

func (rs *Service) CreateRedirectsFromContentserverexport(
	w http.ResponseWriter,
	r *http.Request,
	old, new map[string]*content.RepoNode) error {
	rs.l.Info("CreateRedirectsFromContentserverexport called ")
	return rs.api.CreateRedirects(r.Context(),
		redirectcommand.CreateRedirects{
			OldState: old,
			NewState: new,
		})
}

func (rs *Service) GetRedirects(w http.ResponseWriter, r *http.Request) (*redirectstore.RedirectDefinitions, error) {
	return rs.api.GetRedirects(r.Context())
}

func (rs *Service) Search(w http.ResponseWriter, r *http.Request, dimension, id, path string) (*redirectstore.RedirectDefinition, error) {
	return rs.api.Search(r.Context(), redirectquery.Search{
		ID:     id,
		Source: redirectstore.RedirectSource(path),
	})
}

func (rs *Service) Create(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) error {
	return rs.api.CreateRedirect(r.Context(),
		redirectcommand.CreateRedirect{
			RedirectDefinition: def,
		})
}

func (rs *Service) Delete(w http.ResponseWriter, r *http.Request, path string) error {
	return rs.api.DeleteRedirect(r.Context(),
		redirectcommand.DeleteRedirect{
			Source: redirectstore.RedirectSource(path),
		})
}

func (rs *Service) Update(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) error {
	return rs.api.UpdateRedirect(r.Context(),
		redirectcommand.UpdateRedirect{
			RedirectDefinition: def,
		})
}
