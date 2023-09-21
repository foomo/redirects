package redirect

import (
	"github.com/foomo/contentserver/content"
	"github.com/foomo/redirects/redirection"
)

type Dimension string

type RedirectionService interface {
	CreateRedirectsFromContentserverexport(old, new map[Dimension]*content.RepoNode) error
	Search(dimension, id, path string) ([]*redirection.RedirectDefinition, error)
	Create(def *redirection.RedirectDefinition) error
	Delete(id string) error
	Update(def *redirection.RedirectDefinition) error
}
