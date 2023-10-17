package service

import (
	http "net/http"

	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
)

type RedirectDefinitionService interface {
	CreateRedirectsFromContentserverexport(w http.ResponseWriter, r *http.Request, old, new map[string]*content.RepoNode) error
	Search(w http.ResponseWriter, r *http.Request, dimension, id, path string) (*redirectstore.RedirectDefinition, error)
	Create(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) error
	Delete(w http.ResponseWriter, r *http.Request, path string) error
	Update(w http.ResponseWriter, r *http.Request, def *redirectstore.RedirectDefinition) error
	GetRedirects(w http.ResponseWriter, r *http.Request) (*redirectstore.RedirectDefinitions, error)
}
