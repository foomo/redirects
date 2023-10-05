package redirectdefinition

import (
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

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

	return &consolidatedDef, nil
}
