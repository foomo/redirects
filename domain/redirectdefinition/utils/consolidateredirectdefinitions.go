package redirectdefinitionutils

import (
	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// Consolidate redirect definitions by:
// * Making list for update with new and updated definitions
// * Making list for deleting for definitions with empty target id
// * If target of one is source to another one, consolidate those into one definition to prevent multiple redirections
func ConsolidateRedirectDefinitions(
	l *zap.Logger,
	new []*redirectstore.RedirectDefinition,
	current redirectstore.RedirectDefinitions,
	newNodeMap map[string]*content.RepoNode,
) ([]*redirectstore.RedirectDefinition, []redirectstore.EntityID) {

	upsertRedirectDefinitions := []*redirectstore.RedirectDefinition{}
	deletedIDs := []redirectstore.EntityID{}

	// Step 1:
	// check if for the IDs of the new redirects there is already a redirect in the current
	// state and update the target of the new redirect to the target of the current redirect
	currentRedirectsByID := map[string][]*redirectstore.RedirectDefinition{}
	for _, redirectDefinition := range current {
		_, ok := currentRedirectsByID[redirectDefinition.ContentID]
		if !ok {
			currentRedirectsByID[redirectDefinition.ContentID] = []*redirectstore.RedirectDefinition{}
		}
		currentRedirectsByID[redirectDefinition.ContentID] = append(currentRedirectsByID[redirectDefinition.ContentID], redirectDefinition)
	}

	// iterate over the incoming redirects and add the new redirects to the list that should be upserted
	for _, redirectDefinition := range new {
		upsertRedirectDefinitions = append(upsertRedirectDefinitions, redirectDefinition)
		// check if the ID of the new redirect is already in the current list of redirects
		currentDefinitions, ok := currentRedirectsByID[redirectDefinition.ContentID]
		if ok {
			for _, currentDefinition := range currentDefinitions {
				// if a new redirect points to the same ID as an existing redirect we need to reset
				// the target of the existing redirect to the new target
				if currentDefinition.RedirectionType == redirectstore.Automatic {
					currentDefinition.Target = redirectDefinition.Target
					upsertRedirectDefinitions = append(upsertRedirectDefinitions, currentDefinition)
				}
			}
		}
	}

	// Step 2:
	// handle the case where the target of a redirect is no longer available in the current
	// contentserverexport, in this case the redirect should be deleted
	availableTargets := map[string]struct{}{}
	for _, node := range newNodeMap {
		availableTargets[node.URI] = struct{}{}
	}
	for _, redirectDefinition := range current {
		// we should only handle automatic redirects - manually created redirects
		// might point to URLs that are not handled by contentful and thus might not be part
		// of the new contentserverexport
		if redirectDefinition.RedirectionType == redirectstore.Automatic {
			_, ok := availableTargets[string(redirectDefinition.Target)]
			if !ok {
				deletedIDs = append(deletedIDs, redirectDefinition.ID)
			}
		}
	}

	return upsertRedirectDefinitions, deletedIDs
}
