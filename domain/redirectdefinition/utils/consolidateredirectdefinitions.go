package redirectdefinitionutils

import (
	"github.com/foomo/contentserver/content"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
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
	newDefinitions []*storex.RedirectDefinition,
	currentDefinitions storex.RedirectDefinitions,
	newNodeMap map[string]*content.RepoNode,
) ([]*storex.RedirectDefinition, []storex.EntityID) {
	upserts := make(map[string]*storex.RedirectDefinition)
	deletedIDs := []storex.EntityID{}

	// Step 1: Index current redirects by source
	currentBySource := make(map[storex.RedirectSource]*storex.RedirectDefinition)
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
		if def.RedirectionType != storex.RedirectionTypeAutomatic {
			continue
		}

		// Flatten if the current target is being redirected further
		if next, ok := upserts[string(def.Target)]; ok {
			if def.Source != storex.RedirectSource(next.Target) {
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
	redirect *storex.RedirectDefinition,
	redirectsBySource map[storex.RedirectSource]*storex.RedirectDefinition,
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
	existingRedirect, newRedirect *storex.RedirectDefinition,
	upsertRedirectsMap map[string]*storex.RedirectDefinition,
) {
	if existingRedirect.RedirectionType == storex.RedirectionTypeAutomatic {
		existingRedirect.Target = newRedirect.Target
		upsertRedirectsMap[string(existingRedirect.Source)] = existingRedirect
	}
}

func isRedirectObsolete(
	def *storex.RedirectDefinition,
	upserts map[string]*storex.RedirectDefinition,
	availableTargets map[string]struct{},
	validTargets map[string]struct{},
) bool {
	_, isStillUpserted := upserts[string(def.Source)]
	_, isTargetValid := validTargets[string(def.Target)]
	_, isTargetAvailable := availableTargets[string(def.Target)]

	return !isStillUpserted && !isTargetValid && !isTargetAvailable
}

func mapsToSlice(upsertRedirectsMap map[string]*storex.RedirectDefinition) []*storex.RedirectDefinition {
	upsertRedirectDefinitions := make([]*storex.RedirectDefinition, 0, len(upsertRedirectsMap))
	for _, redirect := range upsertRedirectsMap {
		upsertRedirectDefinitions = append(upsertRedirectDefinitions, redirect)
	}

	return upsertRedirectDefinitions
}

// HasCycle checks if adding Source → Target creates a cyclic redirect
func HasCycle(
	rSource storex.RedirectSource,
	rTarget storex.RedirectTarget,
	redirects map[storex.RedirectSource]*storex.RedirectDefinition,
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
		if nextRedirect, exists := redirects[storex.RedirectSource(target)]; exists {
			target = string(nextRedirect.Target)
		} else {
			return false // No further redirects, no cycle
		}
	}
}
