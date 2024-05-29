package redirectdefinitionutils

import (
	"errors"

	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

var (
	nilError = "calling auto create difference with nil arg"
)

func AutoCreateRedirectDefinitions(l *zap.Logger, old, new *content.RepoNode, dimension redirectstore.Dimension) (redirectstore.RedirectDefinitions, error) {
	l.Info("calling auto create difference between old and new repo node state")
	if old == nil || new == nil {
		l.Error(nilError)
		return nil, errors.New(nilError)
	}
	var redirects = make(redirectstore.RedirectDefinitions)
	var newTree = new
	var generateRedirects func(old, new *content.RepoNode)

	generateRedirects = func(old, new *content.RepoNode) {
		sourceURI := old.URI
		targetURI := new.URI
		if sourceURI != targetURI {
			rd := &redirectstore.RedirectDefinition{
				Source:         redirectstore.RedirectSource(sourceURI),
				Target:         redirectstore.RedirectTarget(targetURI),
				Code:           301,
				RespectParams:  true,
				TransferParams: true,
				Dimension:      dimension,
			}
			redirects[rd.Source] = rd
		}
		for key, oldchild := range old.Nodes {
			if newchild, ok := new.Nodes[key]; ok {
				generateRedirects(oldchild, newchild)
			} else {
				findInNewTree := FindNodeById(newTree, key)
				if findInNewTree != nil {
					generateRedirects(oldchild, findInNewTree)
				} else {
					found := false
					for _, redirect := range redirects {
						if string(redirect.Source) == oldchild.URI {
							found = true
							break
						}
					}
					if !found {
						rd := &redirectstore.RedirectDefinition{
							Source:         redirectstore.RedirectSource(oldchild.URI),
							Target:         "",
							Code:           301,
							RespectParams:  true,
							TransferParams: true,
							Dimension:      dimension,
						}
						redirects[rd.Source] = rd
					}
				}
			}
		}
	}
	// Start generating redirects
	generateRedirects(old, new)
	return redirects, nil
}

// GetAllNodes recursively retrieves all nodes from the tree.
func GetAllNodes(node *content.RepoNode, nodesList []*content.RepoNode) []*content.RepoNode {
	if node == nil {
		return nodesList
	}
	// Add the current node to the list.
	nodesList = append(nodesList, node)
	// Recursively process child nodes.
	for _, child := range node.Nodes {
		nodesList = GetAllNodes(child, nodesList)
	}
	return nodesList
}

func FindNodeById(root *content.RepoNode, id string) *content.RepoNode {
	if root.ID == id {
		return root
	}
	for _, child := range root.Nodes {
		found := FindNodeById(child, id)
		if found != nil {
			return found
		}
	}
	return nil
}
