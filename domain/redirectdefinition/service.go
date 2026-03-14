package redirectdefinition

import (
	"fmt"
	"net/http"
	"time"

	"github.com/foomo/contentserver/content"
	commandx "github.com/foomo/redirects/v2/domain/redirectdefinition/command"
	queryx "github.com/foomo/redirects/v2/domain/redirectdefinition/query"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

type SearchParams struct {
	Locale       string                 `json:"locale"`
	Path         string                 `json:"path"`
	Page         int                    `json:"page"`
	PageSize     int                    `json:"pageSize"`
	RedirectType storex.RedirectionType `json:"type,omitempty"`
	ActiveState  storex.ActiveStateType `json:"activeState,omitempty"`
	Sort         storex.Sort            `json:"sort"`
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
		commandx.CreateRedirects{
			OldState: oldState,
			NewState: newState,
		})
}

// GetRedirects returns all redirects
// internal use only
func (rs *Service) GetRedirects(_ http.ResponseWriter, r *http.Request) (map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition, error) {
	return rs.api.GetRedirects(r.Context())
}

// Search for a redirect
// used by frontend
func (rs *Service) Search(
	_ http.ResponseWriter,
	r *http.Request,
	params *SearchParams,
) (*storex.PaginatedResult, *storex.RedirectDefinitionError) {
	if params.Page < 1 {
		params.Page = 1
	}

	if params.PageSize < 1 {
		params.PageSize = 10 // Default page size
	}

	site, err := rs.api.getSiteIdentifierProvider(r)
	if err != nil {
		return nil, storex.NewRedirectDefinitionError(err.Error())
	}

	result, err := rs.api.Search(r.Context(), queryx.Search{
		Source:       storex.RedirectSource(params.Path),
		Dimension:    storex.Dimension(fmt.Sprintf("%s-%s", site, params.Locale)),
		Page:         params.Page,
		PageSize:     params.PageSize,
		RedirectType: params.RedirectType,
		ActiveState:  params.ActiveState,
		Sort:         params.Sort,
	})
	if err != nil {
		return nil, storex.NewRedirectDefinitionError(err.Error())
	}

	return result, nil
}

// Create a redirect
// used by frontend
func (rs *Service) Create(_ http.ResponseWriter, r *http.Request, def *storex.RedirectDefinition, locale string) (storex.EntityID, *storex.RedirectDefinitionError) {
	site, err := rs.api.getSiteIdentifierProvider(r)
	if err != nil {
		return "", storex.NewRedirectDefinitionError(err.Error())
	}

	def.Dimension = storex.Dimension(fmt.Sprintf("%s-%s", site, locale))
	rs.api.setLastUpdatedBy(r.Context(), def)

	err = rs.api.CreateRedirect(r.Context(),
		commandx.CreateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return "", storex.NewRedirectDefinitionError(err.Error())
	}

	return def.ID, nil
}

// Delete a redirect
// used by frontend
func (rs *Service) Delete(_ http.ResponseWriter, r *http.Request, id string) *storex.RedirectDefinitionError {
	err := rs.api.DeleteRedirect(r.Context(),
		commandx.DeleteRedirect{
			ID: storex.EntityID(id),
		})
	if err != nil {
		return storex.NewRedirectDefinitionError(err.Error())
	}

	return nil
}

// Update a redirect
// used by frontend
func (rs *Service) Update(_ http.ResponseWriter, r *http.Request, def *storex.RedirectDefinition) *storex.RedirectDefinitionError {
	def.Updated = storex.NewDateTime(time.Now())
	rs.api.setLastUpdatedBy(r.Context(), def)

	err := rs.api.UpdateRedirect(r.Context(),
		commandx.UpdateRedirect{
			RedirectDefinition: def,
		})
	if err != nil {
		return storex.NewRedirectDefinitionError(err.Error())
	}

	return nil
}

// UpdateStates updates a redirects state
// used by frontend
func (rs *Service) UpdateStates(_ http.ResponseWriter, r *http.Request, ids []*storex.EntityID, state bool) *storex.RedirectDefinitionError {
	// Fetch all redirects by IDs
	redirects, err := rs.api.repo.FindByIDs(r.Context(), ids)
	if err != nil {
		return storex.NewRedirectDefinitionError("Failed to fetch redirects: " + err.Error())
	}

	// Update each redirect in memory
	for _, def := range redirects {
		def.Stale = !state // flip the value because we are updating the stale field
		def.Updated = storex.NewDateTime(time.Now())
		rs.api.setLastUpdatedBy(r.Context(), def)
	}

	err = rs.api.UpdateRedirectsState(r.Context(), commandx.UpdateRedirectsState{
		RedirectDefinitions: redirects,
	})
	if err != nil {
		return storex.NewRedirectDefinitionError("Failed to update redirects: " + err.Error())
	}

	return nil
}
