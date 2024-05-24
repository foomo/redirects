package redirectdefinition

import (
	"context"
	"net/http"

	"github.com/foomo/contentserver/content"
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

type Service struct {
	l   *zap.Logger
	ctx context.Context
	api *API
}

func NewService(l *zap.Logger, ctx context.Context, api *API) *Service {
	return &Service{
		l:   l,
		ctx: ctx,
		api: api,
	}
}

// CreateRedirectsFromContentserverexport creates redirects from contentserverexport
// internal use only
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

// GetRedirects returns all redirects
// internal use only
func (rs *Service) GetRedirects() (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
	return rs.api.GetRedirects(rs.ctx)
}

// Search for a redirect
// used by frontend
func (rs *Service) Search(w http.ResponseWriter, r *http.Request, dimension, id, path string) (*redirectstore.RedirectDefinitions, *redirectstore.RedirectDefinitionError) {
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

// Delete a redirect
// used by frontend
func (rs *Service) Delete(w http.ResponseWriter, r *http.Request, path, dimension string) *redirectstore.RedirectDefinitionError {
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
