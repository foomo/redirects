package service

import (
	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
)

type RedirectDefinitionService interface {
	CreateRedirectsFromContentserverexport(old, new map[string]*content.RepoNode) error
	Search(dimension, id, path string) (*redirectstore.RedirectDefinition, error)
	Create(def *redirectstore.RedirectDefinition) error
	Delete(id string) error
	Update(def *redirectstore.RedirectDefinition) error
	GetRedirects() (*redirectstore.RedirectDefinitions, error)
}
