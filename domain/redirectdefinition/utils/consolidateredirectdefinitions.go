package redirectdefinitionutils

import (
	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// ConsolidateRedirectDefinitions reconciles new and existing redirect definitions.
//
// This function ensures redirect consistency by performing the following:
//   - Adds or updates automatic redirects from `newDefinitions`.
//   - Flattens redirect chains (e.g., /a → /b → /c becomes /a → /c).
//   - Detects and marks cyclic redirects as stale (e.g., /a → /b → /a).
//   - Marks redirects for deletion if their targets no longer exist in the latest content tree
//     and are not present in the new redirect set.
//   - Ensures existing automatic redirects remain valid, updated, or removed appropriately.
//
// Parameters:
//   - l: Logger for warnings and cycle detection debug info.
//   - newDefinitions: Newly generated redirects (e.g., based on URI changes).
//   - currentDefinitions: Existing stored redirects (typically from the database).
//   - newNodeMap: Latest content state, used to determine available (valid) target URIs.
//
// Returns:
//   - A slice of redirect definitions to upsert (insert or update).
//   - A slice of redirect IDs that are considered obsolete and should be deleted.
func ConsolidateRedirectDefinitions(
	l *zap.Logger,
	newDefinitions []*redirectstore.RedirectDefinition,
	currentDefinitions redirectstore.RedirectDefinitions,
	newNodeMap map[string]*content.RepoNode,
) ([]*redirectstore.RedirectDefinition, []redirectstore.EntityID) {
	upserts := make(map[string]*redirectstore.RedirectDefinition)
	deletedIDs := []redirectstore.EntityID{}

	// Step 1: Index current redirects by source
	currentBySource := make(map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition)
	for _, def := range currentDefinitions {
		currentBySource[def.Source] = def
	}

	// Step 2: Process new redirects
	for _, def := range newDefinitions {
		staleIfCyclic(l, def, currentBySource)
		upserts[string(def.Source)] = def

		if existing, ok := currentBySource[def.Source]; ok {
			updateRedirectTarget(existing, def, upserts)
		}
	}

	// Step 3: Mark targets from content + new redirects as valid
	availableTargets := make(map[string]struct{})
	for _, node := range newNodeMap {
		availableTargets[node.URI] = struct{}{}
	}

	validTargets := make(map[string]struct{})
	for _, def := range newDefinitions {
		validTargets[string(def.Target)] = struct{}{}
	}

	// Step 4: Process old redirects for flattening and cleanup
	for _, def := range currentDefinitions {
		if def.RedirectionType != redirectstore.RedirectionTypeAutomatic {
			continue
		}

		// Flatten if the current target is being redirected further
		if next, ok := upserts[string(def.Target)]; ok {
			if def.Source != redirectstore.RedirectSource(next.Target) {
				def.Target = next.Target
			}
		}

		// Detect cycles after flattening
		staleIfCyclic(l, def, currentBySource)

		// Check if obsolete → delete
		if isRedirectObsolete(def, upserts, availableTargets, validTargets) {
			deletedIDs = append(deletedIDs, def.ID)
			continue
		}

		upserts[string(def.Source)] = def
	}

	return mapsToSlice(upserts), deletedIDs
}

func staleIfCyclic(
	l *zap.Logger,
	redirect *redirectstore.RedirectDefinition,
	redirectsBySource map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition,
) {
	if HasCycle(redirect.Source, redirect.Target, redirectsBySource) {
		redirect.Stale = true
		l.Warn("Cycle detected, marking redirect as stale",
			zap.String("source", string(redirect.Source)),
			zap.String("target", string(redirect.Target)),
		)
	}
}

func updateRedirectTarget(
	existingRedirect, newRedirect *redirectstore.RedirectDefinition,
	upsertRedirectsMap map[string]*redirectstore.RedirectDefinition,
) {
	if existingRedirect.RedirectionType == redirectstore.RedirectionTypeAutomatic {
		existingRedirect.Target = newRedirect.Target
		upsertRedirectsMap[string(existingRedirect.Source)] = existingRedirect
	}
}

func isRedirectObsolete(
	def *redirectstore.RedirectDefinition,
	upserts map[string]*redirectstore.RedirectDefinition,
	availableTargets map[string]struct{},
	validTargets map[string]struct{},
) bool {
	_, isStillUpserted := upserts[string(def.Source)]
	_, isTargetValid := validTargets[string(def.Target)]
	_, isTargetAvailable := availableTargets[string(def.Target)]

	return !isStillUpserted && !isTargetValid && !isTargetAvailable
}

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
