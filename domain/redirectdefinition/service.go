package redirectdefinition

import (
	"net/http"

	"github.com/foomo/contentserver/content"
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

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

// CreateRedirectsFromContentserverexport creates redirects from contentserverexport
// internal use only
func (rs *Service) CreateRedirectsFromContentserverexport(
	_ http.ResponseWriter,
	r *http.Request,
	old,
	new map[string]*content.RepoNode,
) error {
	rs.l.Info("CreateRedirectsFromContentserverexport called ")
	return rs.api.CreateRedirects(r.Context(),
		redirectcommand.CreateRedirects{
			OldState: old,
			NewState: new,
		})
}

// GetRedirects returns all redirects
// internal use only
func (rs *Service) GetRedirects(_ http.ResponseWriter, r *http.Request) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
	return rs.api.GetRedirects(r.Context())
}

// Search for a redirect
// used by frontend
func (rs *Service) Search(_ http.ResponseWriter, r *http.Request, dimension, id, path string) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, *redirectstore.RedirectDefinitionError) {
	result, err := rs.api.Search(r.Context(), redirectquery.Search{
		ID:        id,
		Source:    redirectstore.RedirectSource(path),
		Dimension: redirectstore.Dimension(dimension),
	})
	if err != nil {
		return nil, redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return result, nil
}

// Create a redirect
// used by frontend
func (rs *Service) Create(_ http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) *redirectstore.RedirectDefinitionError {
	err := rs.api.CreateRedirect(r.Context(),
		redirectcommand.CreateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return nil
}

// Delete a redirect
// used by frontend
func (rs *Service) Delete(_ http.ResponseWriter, r *http.Request, path, dimension string) *redirectstore.RedirectDefinitionError {
	err := rs.api.DeleteRedirect(r.Context(),
		redirectcommand.DeleteRedirect{
			Source:    redirectstore.RedirectSource(path),
			Dimension: redirectstore.Dimension(dimension),
		})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return nil
}

// Update a redirect
// used by frontend
func (rs *Service) Update(_ http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) *redirectstore.RedirectDefinitionError {
	err := rs.api.UpdateRedirect(r.Context(),
		redirectcommand.UpdateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return nil
}
