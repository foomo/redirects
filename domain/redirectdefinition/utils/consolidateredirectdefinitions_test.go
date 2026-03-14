package redirectdefinitionutils_test

import (
	"testing"

	"github.com/foomo/contentserver/content"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	utilsx "github.com/foomo/redirects/v2/domain/redirectdefinition/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_ConsolidateRedirectDefinitions(t *testing.T) {
	t.Parallel()

	// Existing content nodes (targets that exist)
	currentNodes := map[string]*content.RepoNode{
		"HMD-de": {ID: "1", URI: "/redirects-test-de-03"},
	}

	// Old redirects before consolidation
	oldRedirects := storex.RedirectDefinitions{
		"/redirects-test-de-01": {
			ID:              "1",
			ContentID:       "1",
			Source:          "/redirects-test-de-01",
			Target:          "/redirects-test-de-02",
			RedirectionType: storex.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
	}

	// New redirects coming into the system
	newRedirects := []*storex.RedirectDefinition{
		{
			ID:              "2",
			ContentID:       "1",
			Source:          "/redirects-test-de-02",
			Target:          "/redirects-test-de-03",
			RedirectionType: storex.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
	}

	// Expected Results
	// `/redirects-test-de-01` should point to `/redirects-test-de-03`
	// `/redirects-test-de-02` should remain as a valid redirect
	expectedUpdated := []*storex.RedirectDefinition{
		{
			ID:              "1",
			ContentID:       "1",
			Source:          "/redirects-test-de-01",
			Target:          "/redirects-test-de-03", // Updated target
			RedirectionType: storex.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
		{
			ID:              "2",
			ContentID:       "1",
			Source:          "/redirects-test-de-02",
			Target:          "/redirects-test-de-03", // New redirect added
			RedirectionType: storex.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
	}

	expectedDeleted := []storex.EntityID{} // No deletions expected

	// Run the function
	updatedDefs, deletedIDs := utilsx.ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)

	// Assertions
	assert.Len(t, updatedDefs, len(expectedUpdated), "Mismatch in updated redirect count")
	assert.Len(t, deletedIDs, len(expectedDeleted), "Mismatch in deleted redirect count")

	// Ensure that expected updates exist
	for _, expected := range expectedUpdated {
		found := false

		for _, actual := range updatedDefs {
			if expected.Source == actual.Source {
				assert.Equal(t, expected.Target, actual.Target, "Unexpected target for source %s", expected.Source)

				found = true

				break
			}
		}

		assert.True(t, found, "Expected redirect not found for source: %s", expected.Source)
	}
}

func Test_ConsolidateRedirectDefinitions_WithCycle(t *testing.T) {
	t.Parallel()

	// Existing redirects with a cycle: /a → /b → /c → /a
	oldRedirects := storex.RedirectDefinitions{
		"/a": {ID: "1", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"},
		"/b": {ID: "2", ContentID: "2", Source: "/b", Target: "/c", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"},
		"/c": {ID: "3", ContentID: "3", Source: "/c", Target: "/a", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"}, // Cycle here
	}

	newRedirects := []*storex.RedirectDefinition{
		{ID: "4", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global", Stale: false},
		{ID: "5", ContentID: "2", Source: "/b", Target: "/c", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global", Stale: false},
		{ID: "6", ContentID: "3", Source: "/c", Target: "/a", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global", Stale: false}, // Cycle here
	}

	currentNodes := map[string]*content.RepoNode{}

	updatedDefs, deletedIDs := utilsx.ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)

	require.Len(t, updatedDefs, 3, "All three redirects should remain in the output")
	assert.Empty(t, deletedIDs, "No redirects should be deleted, only marked as stale")

	expectedStale := map[string]bool{
		"/a": true,
		"/b": true,
		"/c": true,
	}

	for _, def := range updatedDefs {
		assert.True(t, expectedStale[string(def.Source)], "Redirect %s should be marked as stale", def.Source)
		assert.True(t, def.Stale, "Redirect %s should have Stale=true", def.Source)
	}
}

func Test_ConsolidateRedirectDefinitions_NoCycle(t *testing.T) {
	t.Parallel()

	// Existing redirects with a valid sequence: /a → /b → /c → /d
	oldRedirects := storex.RedirectDefinitions{
		"/a": {ID: "1", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"},
		"/b": {ID: "2", ContentID: "2", Source: "/b", Target: "/c", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"},
		"/c": {ID: "3", ContentID: "3", Source: "/c", Target: "/d", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"},
	}

	// New redirects (updated target for `/b`)
	newRedirects := []*storex.RedirectDefinition{
		{ID: "4", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"},
		{ID: "5", ContentID: "2", Source: "/b", Target: "/d", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"}, // Updated target
		{ID: "6", ContentID: "3", Source: "/c", Target: "/d", RedirectionType: storex.RedirectionTypeAutomatic, Dimension: "global"},
	}

	// Available content nodes
	currentNodes := map[string]*content.RepoNode{
		"global": {ID: "1", URI: "/d"},
	}

	// Run the function
	updatedDefs, deletedIDs := utilsx.ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)

	// Assertions
	assert.Len(t, updatedDefs, len(newRedirects), "Mismatch in updated redirect count")
	assert.Empty(t, deletedIDs, "No redirects should be deleted")

	// Ensure the new target for `/b` is correctly updated
	for _, updated := range updatedDefs {
		if updated.Source == "/b" {
			assert.Equal(t, "/d", string(updated.Target), "Redirect target should be updated correctly")
		}
	}
}

func Test_ConsolidateRedirectDefinitions_SkipAndDeleteSelfRedirect(t *testing.T) {
	t.Parallel()

	currentNodes := map[string]*content.RepoNode{
		"HMD-de": {ID: "2", URI: "/herren/bekleidung-neu"},
	}

	// Old redirect: /herren/bekleidung → /herren/bekleidung-neu
	oldRedirects := storex.RedirectDefinitions{
		"/herren/bekleidung": {
			ID:              "1",
			ContentID:       "1",
			Source:          "/herren/bekleidung",
			Target:          "/herren/bekleidung-neu",
			RedirectionType: storex.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
			Stale:           true,
		},
	}

	// New redirect: /herren/bekleidung-neu → /herren/bekleidung (revert)
	newRedirects := []*storex.RedirectDefinition{
		{
			ID:              "2",
			ContentID:       "1",
			Source:          "/herren/bekleidung-neu",
			Target:          "/herren/bekleidung",
			RedirectionType: storex.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
			Stale:           true,
		},
	}

	// Run consolidation
	updatedDefs, deletedIDs := utilsx.ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)

	// Expect both redirects to remain (even if stale), no deletion
	require.Len(t, updatedDefs, 2)

	// Assert both sources are in the result
	expectedSources := map[string]string{
		"/herren/bekleidung":     "/herren/bekleidung-neu",
		"/herren/bekleidung-neu": "/herren/bekleidung",
	}

	for _, def := range updatedDefs {
		expectedTarget, ok := expectedSources[string(def.Source)]
		assert.True(t, ok, "unexpected source in updatedDefs: %s", def.Source)
		assert.Equal(t, expectedTarget, string(def.Target), "unexpected target for %s", def.Source)
		assert.True(t, def.Stale, "redirect should be marked as stale: %s", def.Source)
	}

	// No redirect should be deleted
	assert.Empty(t, deletedIDs, "no redirects should be deleted")
}

// 🔹 Test Cases for HasCycle 🔹

func Test_HasCycle_DetectsCycle(t *testing.T) {
	t.Parallel()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/a"}, // Cycle: A → B → C → A
	}

	assert.True(t, utilsx.HasCycle("/a", "/b", redirects), "Cycle should be detected for /a → /b")
	assert.True(t, utilsx.HasCycle("/b", "/c", redirects), "Cycle should be detected for /b → /c")
	assert.True(t, utilsx.HasCycle("/c", "/a", redirects), "Cycle should be detected for /c → /a")
}

func Test_HasCycle_NoCycle(t *testing.T) {
	t.Parallel()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/d"}, // No cycle
	}

	assert.False(t, utilsx.HasCycle("/a", "/b", redirects), "No cycle should be detected for /a → /b")
	assert.False(t, utilsx.HasCycle("/b", "/c", redirects), "No cycle should be detected for /b → /c")
	assert.False(t, utilsx.HasCycle("/c", "/d", redirects), "No cycle should be detected for /c → /d")
}

func Test_HasCycle_SingleRedirect(t *testing.T) {
	t.Parallel()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
	}

	assert.False(t, utilsx.HasCycle("/a", "/b", redirects), "No cycle should be detected for /a → /b in isolation")
}

func Test_HasCycle_SelfReferencingRedirect(t *testing.T) {
	t.Parallel()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{
		"/a": {Source: "/a", Target: "/a"}, // Self-referencing cycle
	}

	assert.True(t, utilsx.HasCycle("/a", "/a", redirects), "Self-referencing cycle should be detected for /a → /a")
}

func Test_HasCycle_LongChainNoCycle(t *testing.T) {
	t.Parallel()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/d"},
		"/d": {Source: "/d", Target: "/e"},
		"/e": {Source: "/e", Target: "/f"},
	}

	assert.False(t, utilsx.HasCycle("/a", "/b", redirects), "No cycle should be detected for a long chain without cycles")
}

func Test_HasCycle_ComplexCase(t *testing.T) {
	t.Parallel()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/d"},
		"/d": {Source: "/d", Target: "/b"}, // Cycle: B → C → D → B
	}

	assert.True(t, utilsx.HasCycle("/b", "/c", redirects), "Cycle should be detected for /b → /c")
	assert.True(t, utilsx.HasCycle("/c", "/d", redirects), "Cycle should be detected for /c → /d")
	assert.True(t, utilsx.HasCycle("/d", "/b", redirects), "Cycle should be detected for /d → /b")
}

func Test_HasCycle_RepeatingNodeButNoCycle(t *testing.T) {
	t.Parallel()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/d"},
		"/d": {Source: "/d", Target: "/e"},
		"/e": {Source: "/e", Target: "/f"},
		"/x": {Source: "/x", Target: "/b"}, // points into an existing chain
	}

	assert.False(t, utilsx.HasCycle("/x", "/b", redirects))
}
