package redirectdefinition

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/foomo/contentserver/content"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

//go:embed contentNodesTest.json
//var contentNodes []byte

// p := map[string]*content.RepoNode{}
// err := json.Unmarshal(contentNodes, &p)
// if err != nil {
// 	fmt.Println(err)
// }
// pChanged := map[string]*content.RepoNode{}
// err = json.Unmarshal(contentNodesChanged, &pChanged)
// if err != nil {
// 	fmt.Println(err)
// }

func Test_AutoCreateRedirectDefinitions(t *testing.T) {
	old := &content.RepoNode{
		ID:   "1",
		URI:  "/main",
		Name: "Root",
		Nodes: map[string]*content.RepoNode{
			"2": {
				ID:    "2",
				URI:   "/main/herren",
				Name:  "Node2",
				Nodes: nil,
			},
			"3": {
				ID:   "3",
				URI:  "/main/damen/kleidung",
				Name: "Node3",
				Nodes: map[string]*content.RepoNode{
					"4": {
						ID:    "4",
						URI:   "/main/damen/kleidung/schuhe",
						Name:  "Node4",
						Nodes: nil,
					},
					"5": {
						ID:    "5",
						URI:   "/main/damen/kleidung/schuhe-1",
						Name:  "Node5",
						Nodes: nil,
					},
				},
			},
		},
	}
	new := &content.RepoNode{
		ID:   "1",
		URI:  "/main",
		Name: "Root",
		Nodes: map[string]*content.RepoNode{
			"2": {
				ID:    "2",
				URI:   "/main/herren",
				Name:  "Node2",
				Nodes: nil,
			},
			"3": {
				ID:   "3",
				URI:  "/main/damen/kleidung",
				Name: "Node3",
				Nodes: map[string]*content.RepoNode{
					"4": {
						ID:    "4",
						URI:   "/main/damen/kleidung/schuhe-new",
						Name:  "Node4",
						Nodes: nil,
					},
				},
			},
		},
	}
	redirects, err := AutoCreateRedirectDefinitions(zap.L(), old, new)
	if err != nil {
		fmt.Print(err)
	}
	assert.NoError(t, err)
	assert.Equal(t, len(redirects), 2)
}

// FindAllURIs finds and collects all URIs in the tree starting from the given node.
// allUris := FindAllURIs(tree1, []string{})
//
//	for _, path := range allUris {
//		fmt.Println(path)
//	}
func FindAllURIs(node *content.RepoNode, uris []string) []string {
	// Append the current node's URI to the slice of URIs.
	uris = append(uris, node.URI)
	// Recursively find URIs in child nodes.
	for _, child := range node.Nodes {
		uris = FindAllURIs(child, uris)
	}

	return uris
}

// FindAllPaths finds all possible paths by adding URI values.
func FindAllPaths(root *content.RepoNode) []string {
	var result []string
	if len(root.Nodes) == 0 {
		result = append(result, root.URI)
		return result
	}
	for _, child := range root.Nodes {
		result = append(result, FindAllPaths(child)...)
	}
	return result
}

// FindNodeByURI finds a node by its URI in an n-ary tree.
// targetURI := "/main/damen/kleidung/roecke"
// foundNode := FindNodeByURI(tree1, targetURI)
// // Print the result.
//
//	if foundNode != nil {
//		fmt.Printf("Node with URI %s found: ID=%s, Name=%s\n", targetURI, foundNode.ID, foundNode.Name)
//	} else {
//
//		fmt.Printf("Node with URI %s not found.\n", targetURI)
//	}
func FindNodeByURI(root *content.RepoNode, uri string) *content.RepoNode {
	if root.URI == uri {
		return root
	}
	for _, child := range root.Nodes {
		found := FindNodeByURI(child, uri)
		if found != nil {
			return found
		}
	}
	return nil
}
