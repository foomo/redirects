package service

import (
	"net/http"

	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
)

// AdminService is the interface for the admin service
// the service is responsible for the admin endpoints
// will be exposed to the frontend
type AdminService interface {
	Search(w http.ResponseWriter, r *http.Request, dimension, id, path string) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, *redirectstore.RedirectDefinitionError)
	Create(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) *redirectstore.RedirectDefinitionError
	Delete(w http.ResponseWriter, r *http.Request, path, dimension string) *redirectstore.RedirectDefinitionError
	Update(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) *redirectstore.RedirectDefinitionError
}

// InternalService is the interface for the internal service
// the service is responsible for the internal endpoints
// will not be exposed only to other backend services
type InternalService interface {
	CreateRedirectsFromContentserverexport(w http.ResponseWriter, r *http.Request, old, new map[string]*content.RepoNode) error
	GetRedirects() (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error)
}
