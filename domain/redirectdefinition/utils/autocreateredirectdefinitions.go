package redirectdefinitionutils

import (
	"errors"

	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// AutoCreateRedirectDefinitions generates automatic redirects based on the difference between the old and new content tree.
// find new.ID to old.ID and check if the URI is different, if it is different, create a redirect
func AutoCreateRedirectDefinitions(
	_ *zap.Logger,
	oldMap, newMap map[string]*content.RepoNode,
	dimension redirectstore.Dimension,
) ([]*redirectstore.RedirectDefinition, error) {
	if len(oldMap) == 0 || len(newMap) == 0 {
		return nil, errors.New("calling auto create difference with nil arguments")
	}
	redirects := []*redirectstore.RedirectDefinition{}

	for newNodeID, newNode := range newMap {
		oldNode, ok := oldMap[newNodeID]
		if ok {
			if oldNode.URI != newNode.URI {
				rd := &redirectstore.RedirectDefinition{
					ID:              redirectstore.NewEntityID(),
					ContentID:       newNodeID,
					Source:          redirectstore.RedirectSource(oldNode.URI),
					Target:          redirectstore.RedirectTarget(newNode.URI),
					Code:            301,
					RespectParams:   true,
					TransferParams:  true,
					RedirectionType: redirectstore.Automatic,
					Dimension:       dimension,
				}
				redirects = append(redirects, rd)
			}
		}
	}

	return redirects, nil
}

// CreateFlatRepoNodeMap recursively retrieves all nodes from the tree and returns them in a flat map.
func CreateFlatRepoNodeMap(node *content.RepoNode, nodeMap map[string]*content.RepoNode) map[string]*content.RepoNode {
	if node == nil {
		return nodeMap
	}
	// Add the current node to the list.
	nodeMap[node.ID] = node
	// Recursively process child nodes.
	for _, child := range node.Nodes {
		nodeMap = CreateFlatRepoNodeMap(child, nodeMap)
	}
	return nodeMap
}
