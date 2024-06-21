package redirectdefinition

import (
	"fmt"
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

	enableCreationOfAutomaticRedirects enabledFunc
}

func NewService(l *zap.Logger, api *API, options ...ServiceOption) *Service {
	s := &Service{
		l:                                  l,
		api:                                api,
		enableCreationOfAutomaticRedirects: defaultEnabledFunc,
	}

	for _, o := range options {
		o(s)
	}

	return s
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
	if !rs.enableCreationOfAutomaticRedirects() {
		rs.l.Info("CreateRedirectsFromContentserverexport not enabled")
		return nil
	}
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
func (rs *Service) Search(_ http.ResponseWriter, r *http.Request, locale, path string) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, *redirectstore.RedirectDefinitionError) {
	site, err := rs.api.getSiteIdentifierProvider(r)
	if err != nil {
		return nil, redirectstore.NewRedirectDefinitionError(err.Error())
	}

	result, err := rs.api.Search(r.Context(), redirectquery.Search{
		Source:    redirectstore.RedirectSource(path),
		Dimension: redirectstore.Dimension(fmt.Sprintf("%s-%s", site, locale)),
	})
	if err != nil {
		return nil, redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return result, nil
}

// Create a redirect
// used by frontend
func (rs *Service) Create(_ http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition, locale string) (redirectstore.EntityID, *redirectstore.RedirectDefinitionError) {
	site, err := rs.api.getSiteIdentifierProvider(r)
	if err != nil {
		return "", redirectstore.NewRedirectDefinitionError(err.Error())
	}
	def.Dimension = redirectstore.Dimension(fmt.Sprintf("%s-%s", site, locale))

	err = rs.api.CreateRedirect(r.Context(),
		redirectcommand.CreateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return "", redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return def.ID, nil
}

// Delete a redirect
// used by frontend
func (rs *Service) Delete(_ http.ResponseWriter, r *http.Request, id string) *redirectstore.RedirectDefinitionError {
	err := rs.api.DeleteRedirect(r.Context(),
		redirectcommand.DeleteRedirect{
			ID: redirectstore.EntityID(id),
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
