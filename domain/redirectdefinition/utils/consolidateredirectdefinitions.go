package redirectdefinitionutils

import (
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// Consolidate redirect definitions by:
// * Removing ones that have empty target definition
// * Removing ones that don't exist in new definitions
// * If target of one is source to another one, consolidate those into one to prevent multiple redirections
func ConsolidateRedirectDefinitions(
	l *zap.Logger,
	old, new redirectstore.RedirectDefinitions,
) (*redirectstore.RedirectDefinitions, error) {

	consolidatedDef := make(redirectstore.RedirectDefinitions)

	// Copy new definitions to the consolidated map
	for source, definition := range new {
		// If Target is empty in new definitions, skip it
		if definition.Target != "" {
			consolidatedDef[source] = definition
		}
	}

	// Remove definitions from the consolidated map if they exist in old but not in new
	for source := range old {
		if _, found := consolidatedDef[source]; !found {
			// Definition exists in old but not in new, remove it
			delete(old, source)
		}
	}

	// Check for circular references and update the targets if needed
	for _, definition := range consolidatedDef {
		target := definition.Target
		for {
			if nextDefinition, found := consolidatedDef[redirectstore.RedirectSource(target)]; found {
				// If the target is also a source in another definition, update the target
				if nextDefinition.Target != target {
					definition.Target = nextDefinition.Target

					// Circular reference detected, remove the target
					delete(consolidatedDef, redirectstore.RedirectSource(target))
					break
				}
			} else {
				// No more references found, update the target
				definition.Target = target
				break
			}
		}
	}
	return &consolidatedDef, nil
}
