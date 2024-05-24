package redirectdefinitionutils

import (
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// Consolidate redirect definitions by:
// * Making list for update with new and updated definitions
// * Making list for deleting for definitions with empty target id
// * If target of one is source to another one, consolidate those into one definition to prevent multiple redirections
func ConsolidateRedirectDefinitions(
	l *zap.Logger,
	new redirectstore.RedirectDefinitions,
	dimension redirectstore.Dimension,
) (updatedDefs redirectstore.RedirectDefinitions, deletedSources []redirectstore.RedirectSource) {

	updatedDefs = make(redirectstore.RedirectDefinitions)

	// Copy new definitions to the consolidated map
	for source, definitionByDimension := range new {
		for _, definition := range definitionByDimension {
			// If Target is empty in new definitions, delete it
			if definition.Target == "" {
				deletedSources = append(deletedSources, source)
			} else {
				if updatedDefs[source] == nil {
					updatedDefs[source] = make(map[redirectstore.Dimension]*redirectstore.RedirectDefinition)
				}
				updatedDefs[source][dimension] = definition
			}
		}
	}

	// Check for circular references and update the targets if needed

	for _, definitionByDimension := range updatedDefs {
		for _, definition := range definitionByDimension {

			target := definition.Target
			for {
				if nextDefinition, found := updatedDefs[redirectstore.RedirectSource(target)]; found {
					// If the target is also a source in another definition, update the target
					if nextDefinition[dimension].Target != target {
						definition.Target = nextDefinition[dimension].Target

						// Circular reference detected, remove the target
						delete(updatedDefs, redirectstore.RedirectSource(target))

						deletedSources = append(deletedSources, redirectstore.RedirectSource(target))
						break
					}
				} else {
					// No more references found, update the target
					definition.Target = target
					break
				}
			}
		}
	}
	return updatedDefs, deletedSources
}
