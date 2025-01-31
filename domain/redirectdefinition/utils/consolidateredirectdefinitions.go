package redirectdefinitionutils

import (
	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// Consolidate redirect definitions by:
// * Making a list for update with new and updated definitions
// * Prevent multiple redirections by consolidating redirects into a single definition.
// * Detecting and marking cyclic redirects as `stale = true`
// * Deleting redirects whose target does not exist in `newNodeMap`
// * Processing both active and inactive redirects, because automatic can be reactivated by user
func ConsolidateRedirectDefinitions(
	l *zap.Logger,
	newDefinitions []*redirectstore.RedirectDefinition,
	current redirectstore.RedirectDefinitions,
	newNodeMap map[string]*content.RepoNode,
) ([]*redirectstore.RedirectDefinition, []redirectstore.EntityID) {
	upsertRedirectDefinitions := []*redirectstore.RedirectDefinition{}
	deletedIDs := []redirectstore.EntityID{}

	// Step 1: Map all redirects by source
	redirectsBySource := make(map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition)
	for _, redirectDefinition := range current {
		redirectsBySource[redirectDefinition.Source] = redirectDefinition
	}

	// Step 2: Process new redirects
	for _, newRedirect := range newDefinitions {
		// Detect cycles and mark `stale = true` instead of deleting
		if HasCycle(newRedirect.Source, newRedirect.Target, redirectsBySource) {
			newRedirect.Stale = true
			l.Warn("Cycle detected, marking redirect as stale",
				zap.String("source", string(newRedirect.Source)),
				zap.String("target", string(newRedirect.Target)),
			)
		}

		upsertRedirectDefinitions = append(upsertRedirectDefinitions, newRedirect)

		// If an existing redirect exists, update its target
		if existingRedirect, exists := redirectsBySource[newRedirect.Source]; exists {
			if existingRedirect.RedirectionType == redirectstore.RedirectionTypeAutomatic {
				// Propagate updates: if an old redirect points to this new source, update its target
				for _, redirect := range redirectsBySource {
					if string(redirect.Target) == string(existingRedirect.Source) {
						redirect.Target = newRedirect.Target
						upsertRedirectDefinitions = append(upsertRedirectDefinitions, redirect)

						l.Info("Updated chained redirect",
							zap.String("source", string(redirect.Source)),
							zap.String("new_target", string(redirect.Target)),
						)
					}
				}

				existingRedirect.Target = newRedirect.Target
				upsertRedirectDefinitions = append(upsertRedirectDefinitions, existingRedirect)

				l.Info("Updated automatic redirect target",
					zap.String("source", string(existingRedirect.Source)),
					zap.String("new_target", string(existingRedirect.Target)),
				)
			}
		}
	}

	// Step 3: Identify redirects whose targets are no longer available in the current ContentServerExport.
	availableTargets := make(map[string]struct{})
	for _, node := range newNodeMap {
		availableTargets[node.URI] = struct{}{}
	}

	for _, redirectDefinition := range current {
		// Process only automatic redirects.
		// Manually created redirects might point to external URLs or custom cases
		// that are not covered by the ContentServerExport, so we leave them as is.
		if redirectDefinition.RedirectionType == redirectstore.RedirectionTypeAutomatic {
			if _, exists := availableTargets[string(redirectDefinition.Target)]; !exists {
				// Mark redirect for deletion if its target no longer exists
				deletedIDs = append(deletedIDs, redirectDefinition.ID)

				l.Warn("Redirect target no longer exists, marking for deletion",
					zap.String("source", string(redirectDefinition.Source)),
					zap.String("target", string(redirectDefinition.Target)),
					zap.String("redirect_id", string(redirectDefinition.ID)),
				)
			}
		}
	}

	return upsertRedirectDefinitions, deletedIDs
}

// HasCycle checks if adding Source → Target creates a cyclic redirect
func HasCycle(
	rSource redirectstore.RedirectSource,
	rTarget redirectstore.RedirectTarget,
	redirects map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition,
) bool {
	visited := make(map[string]struct{})
	source := string(rSource)
	target := string(rTarget)

	// Follow the chain of redirects
	for {
		// If we reach an empty target, there is no cycle
		if target == "" {
			return false
		}

		// If the target points back to the original source, we have a cycle
		if target == source {
			return true
		}

		// If we already visited this target, break the cycle
		if _, exists := visited[target]; exists {
			return true
		}

		// Mark this target as visited
		visited[target] = struct{}{}

		// Move to the next redirect in the chain
		if nextRedirect, exists := redirects[redirectstore.RedirectSource(target)]; exists {
			target = string(nextRedirect.Target)
		} else {
			return false // No further redirects, no cycle
		}
	}
}
