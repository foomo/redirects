package redirectdefinition

import (
	"fmt"
	"net/http"
	"time"

	"github.com/foomo/contentserver/content"
	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

type SearchParams struct {
	Locale       string                        `json:"locale"`
	Path         string                        `json:"path"`
	Page         int                           `json:"page"`
	PageSize     int                           `json:"pageSize"`
	RedirectType redirectstore.RedirectionType `json:"type,omitempty"`
	ActiveState  redirectstore.ActiveStateType `json:"activeState,omitempty"`
	Sort         redirectrepository.Sort       `json:"sort"`
}

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
	oldState,
	newState map[string]*content.RepoNode,
) error {
	rs.l.Info("CreateRedirectsFromContentserverexport called ")
	if !rs.enableCreationOfAutomaticRedirects() {
		rs.l.Info("CreateRedirectsFromContentserverexport not enabled")
		return nil
	}
	return rs.api.CreateRedirects(r.Context(),
		redirectcommand.CreateRedirects{
			OldState: oldState,
			NewState: newState,
		})
}

// GetRedirects returns all redirects
// internal use only
func (rs *Service) GetRedirects(_ http.ResponseWriter, r *http.Request) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
	return rs.api.GetRedirects(r.Context())
}

// Search for a redirect
// used by frontend
func (rs *Service) Search(
	_ http.ResponseWriter,
	r *http.Request,
	params *SearchParams,
) (*redirectrepository.PaginatedResult, *redirectstore.RedirectDefinitionError) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10 // Default page size
	}

	site, err := rs.api.getSiteIdentifierProvider(r)
	if err != nil {
		return nil, redirectstore.NewRedirectDefinitionError(err.Error())
	}

	result, err := rs.api.Search(r.Context(), redirectquery.Search{
		Source:       redirectstore.RedirectSource(params.Path),
		Dimension:    redirectstore.Dimension(fmt.Sprintf("%s-%s", site, params.Locale)),
		Page:         params.Page,
		PageSize:     params.PageSize,
		RedirectType: params.RedirectType,
		ActiveState:  params.ActiveState,
		Sort:         params.Sort,
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
	rs.api.setLastUpdatedBy(r.Context(), def)
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
	def.Updated = redirectstore.NewDateTime(time.Now())
	rs.api.setLastUpdatedBy(r.Context(), def)
	err := rs.api.UpdateRedirect(r.Context(),
		redirectcommand.UpdateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError(err.Error())
	}
	return nil
}

// Update a redirects state
// used by frontend
func (rs *Service) UpdateStates(_ http.ResponseWriter, r *http.Request, ids []*redirectstore.EntityID, state bool) *redirectstore.RedirectDefinitionError {
	// Fetch all redirects by IDs
	redirects, err := rs.api.repo.FindByIDs(r.Context(), ids)
	if err != nil {
		return redirectstore.NewRedirectDefinitionError("Failed to fetch redirects: " + err.Error())
	}

	// Update each redirect in memory
	for _, def := range redirects {
		def.Stale = !state // flip the value because we are updating the stale field
		def.Updated = redirectstore.NewDateTime(time.Now())
		rs.api.setLastUpdatedBy(r.Context(), def)
	}

	err = rs.api.UpdateRedirectsState(r.Context(), redirectcommand.UpdateRedirectsState{
		RedirectDefinitions: redirects,
	})
	if err != nil {
		return redirectstore.NewRedirectDefinitionError("Failed to update redirects: " + err.Error())
	}

	return nil
}
