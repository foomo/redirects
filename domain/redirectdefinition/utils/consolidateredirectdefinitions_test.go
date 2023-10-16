package redirectdefinitionutils

import (
	"fmt"
	"testing"

	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_ConsolidateRedirectDefinitions(t *testing.T) {
	old := redirectstore.RedirectDefinitions{
		"damen":  {ID: "1", Source: "damen", Target: "damenish"},
		"her":    {ID: "2", Source: "her", Target: "heren"}, // test if doesn't exist in new it will be removed
		"kinder": {ID: "3", Source: "kinder", Target: "kids"},
	}

	new := redirectstore.RedirectDefinitions{
		"damen":  {ID: "1", Source: "damen", Target: ""}, // test that if target is empty it will be removed
		"kinder": {ID: "3", Source: "kinder", Target: "kids"},
		"kids":   {ID: "3", Source: "kids", Target: "new-kinder"},
		// TODO: Dragana currently works with only 2 circular references, make it work with multiple
		//"new-kinder": {ID: "3", Source: "new-kinder", Target: "newest-kinder"}, // test that if a target is source in another definition it will be consolidated
		"tachen": {ID: "4", Source: "tachen", Target: "new-tachen"}, // test that newly added will be actually added
	}

	updatedExpected := redirectstore.RedirectDefinitions{
		"kinder": {ID: "3", Source: "kinder", Target: "new-kinder"},
		"tachen": {ID: "4", Source: "tachen", Target: "new-tachen"},
	}

	deletedExpected := []redirectstore.RedirectSource{
		"kids",
		"damen",
	}
	updatedDefs, deletedSources := ConsolidateRedirectDefinitions(zap.L(), old, new)
	fmt.Print(deletedSources)
	assert.Equal(t, len(updatedExpected), len(updatedDefs))
	assert.Equal(t, len(deletedExpected), len(deletedSources))

	//make sure that consolidated definitions exist in expected
	for source, definition := range updatedDefs {
		assert.NotNil(t, updatedDefs[source])
		assert.Equal(t, definition, updatedDefs[source])
	}
}
