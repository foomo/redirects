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
		"1": {ID: "1", Source: "damen", Target: "heren"},
		"2": {ID: "2", Source: "kinder", Target: "kids"},
	}

	new := redirectstore.RedirectDefinitions{
		"1": {ID: "1", Source: "damen", Target: ""},
		"3": {ID: "3", Source: "tachen", Target: "new-tachen"},
	}

	consolidatedExpected := redirectstore.RedirectDefinitions{
		"3": {ID: "3", Source: "tachen", Target: "new-tachen"},
	}
	consolidated, err := ConsolidateRedirectDefinitions(zap.L(), old, new)
	if err != nil {
		fmt.Print(err)
	}
	assert.NoError(t, err)
	assert.Equal(t, len(consolidatedExpected), len(*consolidated))

	//make sure that consolidated definitions exist in expected
	for id := range *consolidated {
		assert.NotNil(t, consolidatedExpected[id])
	}
}
