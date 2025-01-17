package service

import (
	"net/http"

	"github.com/foomo/contentserver/content"
	"github.com/foomo/redirects/domain/redirectdefinition"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
)

// AdminService is the interface for the admin service
// the service is responsible for the admin endpoints
// will be exposed to the frontend
type AdminService interface {
	Search(w http.ResponseWriter, r *http.Request, params *redirectdefinition.SearchParams) (*redirectrepository.PaginatedResult, *redirectstore.RedirectDefinitionError)
	Create(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition, locale string, user string) (redirectstore.EntityID, *redirectstore.RedirectDefinitionError)
	Delete(w http.ResponseWriter, r *http.Request, id string) *redirectstore.RedirectDefinitionError
	Update(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition, user string) *redirectstore.RedirectDefinitionError
}

// InternalService is the interface for the internal service
// the service is responsible for the internal endpoints
// will not be exposed only to other backend services
type InternalService interface {
	CreateRedirectsFromContentserverexport(w http.ResponseWriter, r *http.Request, oldState, newState map[string]*content.RepoNode) error
	GetRedirects(w http.ResponseWriter, r *http.Request) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error)
}
