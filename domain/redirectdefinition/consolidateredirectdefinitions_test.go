package redirectdefinition

import (
	"fmt"
	"testing"

	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_ConsolidateRedirectDefinitions(t *testing.T) {
	old := redirectstore.RedirectDefinitions{
		"damen":     {ID: "1", Source: "damen", Target: "damen-new"},
		"damen-new": {ID: "1", Source: "damen-new", Target: "heren"},
		"kinder":    {ID: "2", Source: "kinder", Target: "kids"},
	}

	new := redirectstore.RedirectDefinitions{
		"damen":     {ID: "1", Source: "damen", Target: ""},
		"damen-new": {ID: "1", Source: "damen-new", Target: "heren"},
		"tachen":    {ID: "3", Source: "tachen", Target: "new-tachen"},
	}

	consolidatedExpected := redirectstore.RedirectDefinitions{
		"damen-new": {ID: "1", Source: "damen-new", Target: "heren"},
		"tachen":    {ID: "3", Source: "tachen", Target: "new-tachen"},
	}
	consolidated, err := ConsolidateRedirectDefinitions(zap.L(), old, new)
	if err != nil {
		fmt.Print(err)
	}
	assert.NoError(t, err)
	assert.Equal(t, len(consolidatedExpected), len(*consolidated))

	//make sure that consolidated definitions exist in expected
	for source, definition := range *consolidated {
		assert.NotNil(t, consolidatedExpected[source])
		assert.Equal(t, definition, consolidatedExpected[source])
	}
}
