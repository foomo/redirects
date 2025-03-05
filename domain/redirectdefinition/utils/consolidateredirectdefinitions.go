package redirectdefinitionutils

import (
	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// ConsolidateRedirectDefinitions processes and updates redirect definitions by:
// - Updating existing redirects to their correct targets.
// - Consolidating chains to prevent multiple unnecessary redirections.
// - Detecting and marking cycles (`stale = true`) to avoid infinite loops.
// - Deleting redirects if their target no longer exists in `newNodeMap`.
// - Processing both active and inactive redirects, since redirects can be reactivated later manually.
//
// Parameters:
// - l: Logger for debug/warning messages.
// - newDefinitions: New incoming redirects to be added or updated.
// - current: Existing redirects that need to be consolidated.
// - newNodeMap: Available content nodes (determines if a target is still valid).
//
// Returns:
// - A list of updated redirect definitions (merged and adjusted).
// - A list of entity IDs for redirects that should be deleted.
func ConsolidateRedirectDefinitions(
	l *zap.Logger,
	newDefinitions []*redirectstore.RedirectDefinition,
	current redirectstore.RedirectDefinitions,
	newNodeMap map[string]*content.RepoNode,
) ([]*redirectstore.RedirectDefinition, []redirectstore.EntityID) {
	upsertRedirectsMap := make(map[string]*redirectstore.RedirectDefinition)
	deletedIDs := []redirectstore.EntityID{}

	// Step 1: Map all redirects by source
	redirectsBySource := make(map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition)
	for _, redirectDefinition := range current {
		redirectsBySource[redirectDefinition.Source] = redirectDefinition
	}

	// Step 2: Process new redirects
	for _, newRedirect := range newDefinitions {
		handleCycleCheck(l, newRedirect, redirectsBySource)

		// Store unique redirects
		upsertRedirectsMap[string(newRedirect.Source)] = newRedirect

		// If an existing redirect exists, update its target
		if existingRedirect, exists := redirectsBySource[newRedirect.Source]; exists {
			updateRedirectTarget(existingRedirect, newRedirect, upsertRedirectsMap)
		}
	}

	// Step 3: Identify redirects whose targets are no longer available
	availableTargets := mapAvailableTargets(newNodeMap)
	validTargets := mapValidTargets(newDefinitions)

	// Process old redirects
	for _, redirectDefinition := range current {
		if redirectDefinition.RedirectionType == redirectstore.RedirectionTypeAutomatic {
			// Ensure propagation of target updates
			if newTarget, exists := upsertRedirectsMap[string(redirectDefinition.Target)]; exists {
				redirectDefinition.Target = newTarget.Target
				upsertRedirectsMap[string(redirectDefinition.Source)] = redirectDefinition
			}

			// Detect cycles
			if handleCycleCheck(l, redirectDefinition, redirectsBySource) {
				upsertRedirectsMap[string(redirectDefinition.Source)] = redirectDefinition
				continue
			}

			// Check if redirect should be deleted
			if shouldDeleteRedirect(redirectDefinition, availableTargets, validTargets, upsertRedirectsMap) {
				deletedIDs = append(deletedIDs, redirectDefinition.ID)
			} else {
				upsertRedirectsMap[string(redirectDefinition.Source)] = redirectDefinition
			}
		}
	}

	// Convert map to slice
	upsertRedirectDefinitions := mapsToSlice(upsertRedirectsMap)

	return upsertRedirectDefinitions, deletedIDs
}

// handleCycleCheck detects cyclic redirects and marks them as stale
func handleCycleCheck(
	l *zap.Logger,
	redirect *redirectstore.RedirectDefinition,
	redirectsBySource map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition,
) bool {
	if HasCycle(redirect.Source, redirect.Target, redirectsBySource) {
		redirect.Stale = true
		l.Warn("Cycle detected, marking redirect as stale",
			zap.String("source", string(redirect.Source)),
			zap.String("target", string(redirect.Target)),
		)
		return true
	}
	return false
}

// updateRedirectTarget updates the target of an existing automatic redirect
func updateRedirectTarget(
	existingRedirect, newRedirect *redirectstore.RedirectDefinition,
	upsertRedirectsMap map[string]*redirectstore.RedirectDefinition,
) {
	if existingRedirect.RedirectionType == redirectstore.RedirectionTypeAutomatic {
		existingRedirect.Target = newRedirect.Target
		upsertRedirectsMap[string(existingRedirect.Source)] = existingRedirect
	}
}

// mapAvailableTargets creates a set of available targets from newNodeMap
func mapAvailableTargets(newNodeMap map[string]*content.RepoNode) map[string]struct{} {
	availableTargets := make(map[string]struct{})
	for _, node := range newNodeMap {
		availableTargets[node.URI] = struct{}{}
	}
	return availableTargets
}

// mapValidTargets creates a set of valid targets from newDefinitions
func mapValidTargets(newDefinitions []*redirectstore.RedirectDefinition) map[string]struct{} {
	validTargets := make(map[string]struct{})
	for _, newRedirect := range newDefinitions {
		validTargets[string(newRedirect.Target)] = struct{}{}
	}
	return validTargets
}

// shouldDeleteRedirect determines whether a redirect should be deleted
func shouldDeleteRedirect(
	redirectDefinition *redirectstore.RedirectDefinition,
	availableTargets, validTargets map[string]struct{},
	upsertRedirectsMap map[string]*redirectstore.RedirectDefinition,
) bool {
	_, existsInAvailableTargets := availableTargets[string(redirectDefinition.Target)]
	_, existsInValidTargets := validTargets[string(redirectDefinition.Target)]

	return !existsInAvailableTargets && !existsInValidTargets && upsertRedirectsMap[string(redirectDefinition.Source)] == nil
}

// mapsToSlice converts a map of redirects to a slice
func mapsToSlice(upsertRedirectsMap map[string]*redirectstore.RedirectDefinition) []*redirectstore.RedirectDefinition {
	upsertRedirectDefinitions := make([]*redirectstore.RedirectDefinition, 0, len(upsertRedirectsMap))
	for _, redirect := range upsertRedirectsMap {
		upsertRedirectDefinitions = append(upsertRedirectDefinitions, redirect)
	}
	return upsertRedirectDefinitions
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
