package redirectdefinition

import (
	"net/http"

	"github.com/foomo/contentserver/content"
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// Potentially add Nats to service (still not sure)
type Service struct {
	l   *zap.Logger
	api *API
}

func NewService(l *zap.Logger, api *API) *Service {
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

func (rs *Service) Search(w http.ResponseWriter, r *http.Request, dimension, id, path string) (*redirectstore.RedirectDefinitions, *redirectstore.RedirectDefinitionError) {
	result, err := rs.api.Search(r.Context(), redirectquery.Search{
		ID:     id,
		Source: redirectstore.RedirectSource(path),
	})
	if err != nil {
		return nil, redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return result, nil
}

func (rs *Service) Create(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) *redirectstore.RedirectDefinitionError {
	err := rs.api.CreateRedirect(r.Context(),
		redirectcommand.CreateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return nil
}

func (rs *Service) Delete(w http.ResponseWriter, r *http.Request, path string) *redirectstore.RedirectDefinitionError {
	err := rs.api.DeleteRedirect(r.Context(),
		redirectcommand.DeleteRedirect{
			Source: redirectstore.RedirectSource(path),
		})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return nil
}

func (rs *Service) Update(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) *redirectstore.RedirectDefinitionError {
	err := rs.api.UpdateRedirect(r.Context(),
		redirectcommand.UpdateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return nil
}
