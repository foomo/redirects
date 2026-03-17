package service

import (
	"net/http"

	"github.com/foomo/contentserver/content"
	redirectdefinitionx "github.com/foomo/redirects/v2/domain/redirectdefinition"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
)

// AdminService is the interface for the admin service
// the service is responsible for the admin endpoints
// will be exposed to the frontend
type AdminService interface {
	Search(w http.ResponseWriter, r *http.Request, params *redirectdefinitionx.SearchParams) (*storex.PaginatedResult, *storex.RedirectDefinitionError)
	Create(w http.ResponseWriter, r *http.Request, def *storex.RedirectDefinition, locale string) (storex.EntityID, *storex.RedirectDefinitionError)
	Delete(w http.ResponseWriter, r *http.Request, id string) *storex.RedirectDefinitionError
	Update(w http.ResponseWriter, r *http.Request, def *storex.RedirectDefinition) *storex.RedirectDefinitionError
	UpdateStates(w http.ResponseWriter, r *http.Request, ids []*storex.EntityID, state bool) *storex.RedirectDefinitionError
}

// InternalService is the interface for the internal service
// the service is responsible for the internal endpoints
// will not be exposed only to other backend services
type InternalService interface {
	CreateRedirectsFromContentserverexport(w http.ResponseWriter, r *http.Request, oldState, newState map[string]*content.RepoNode) error
	GetRedirects(w http.ResponseWriter, r *http.Request) (map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition, error)
}
