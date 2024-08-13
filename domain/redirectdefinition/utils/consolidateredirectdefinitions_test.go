package redirectdefinitionutils

import (
	"testing"

	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_ConsolidateRedirectDefinitions(t *testing.T) {
	currentNodes := map[string]*content.RepoNode{
		"HMD-de": {ID: "1", URI: "/redirects-test-de-03"},
	}

	oldRedirects := redirectstore.RedirectDefinitions{
		"/redirects-test-de-01": {ID: "1", ContentID: "1", Source: "/redirects-test-de-01", Target: "/redirects-test-de-02", RedirectionType: redirectstore.Automatic},
	}

	newRedirects := []*redirectstore.RedirectDefinition{
		{ID: "2", ContentID: "1", Source: "/redirects-test-de-02", Target: "/redirects-test-de-03", RedirectionType: redirectstore.Automatic},
	}

	updatedExpected := redirectstore.RedirectDefinitions{
		"/redirects-test-de-01": {ID: "1", ContentID: "1", Source: "/redirects-test-de-01", Target: "/redirects-test-de-03", RedirectionType: redirectstore.Automatic},
		"/redirects-test-de-02": {ID: "2", ContentID: "1", Source: "/redirects-test-de-02", Target: "/redirects-test-de-03", RedirectionType: redirectstore.Automatic},
	}

	deletedExpected := []redirectstore.RedirectSource{}

	updatedDefs, deletedSources := ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)
	assert.Equal(t, len(updatedExpected), len(updatedDefs))
	assert.Equal(t, len(deletedExpected), len(deletedSources))

	//make sure that consolidated definitions exist in expected
	for source, definition := range updatedDefs {
		assert.NotNil(t, updatedDefs[source])
		assert.Equal(t, definition, updatedDefs[source])
	}
}
