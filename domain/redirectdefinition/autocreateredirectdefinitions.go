package redirectdefinition

import (
	"fmt"

	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

func AutoCreateRedirectDefinitions(l *zap.Logger, old, new *content.RepoNode) ([]redirectstore.RedirectDefinition, error) {
	l.Info("calling auto create difference between old and new repo node state")
	var redirects []redirectstore.RedirectDefinition
	var generateRedirects func(old, new *content.RepoNode)
	generateRedirects = func(old, new *content.RepoNode) {
		sourceURI := old.URI
		targetURI := new.URI
		if sourceURI != targetURI {
			redirects = append(redirects, redirectstore.RedirectDefinition{Source: sourceURI, Target: targetURI})
		}
		for key, oldchild := range old.Nodes {
			if newchild, ok := new.Nodes[key]; ok {
				generateRedirects(oldchild, newchild)
			} else {
				fmt.Println("not have", key)
			}
		}
		if len(new.Nodes) < len(old.Nodes) {
			for key, oldchild := range old.Nodes {
				if _, ok := new.Nodes[key]; !ok {
					redirects = append(redirects, redirectstore.RedirectDefinition{Source: oldchild.URI, Target: ""})
				}
			}
		}
	}
	// Start generating redirects
	generateRedirects(old, new)
	return redirects, nil
}
