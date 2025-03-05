package redirectdefinitionutils_test

import (
	"testing"

	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	rdutils "github.com/foomo/redirects/domain/redirectdefinition/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_ConsolidateRedirectDefinitions(t *testing.T) {
	// Existing content nodes (targets that exist)
	currentNodes := map[string]*content.RepoNode{
		"HMD-de": {ID: "1", URI: "/redirects-test-de-03"},
	}

	// Old redirects before consolidation
	oldRedirects := redirectstore.RedirectDefinitions{
		"/redirects-test-de-01": {
			ID:              "1",
			ContentID:       "1",
			Source:          "/redirects-test-de-01",
			Target:          "/redirects-test-de-02",
			RedirectionType: redirectstore.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
	}

	// New redirects coming into the system
	newRedirects := []*redirectstore.RedirectDefinition{
		{
			ID:              "2",
			ContentID:       "1",
			Source:          "/redirects-test-de-02",
			Target:          "/redirects-test-de-03",
			RedirectionType: redirectstore.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
	}

	// Expected Results
	// `/redirects-test-de-01` should point to `/redirects-test-de-03`
	// `/redirects-test-de-02` should remain as a valid redirect
	expectedUpdated := []*redirectstore.RedirectDefinition{
		{
			ID:              "1",
			ContentID:       "1",
			Source:          "/redirects-test-de-01",
			Target:          "/redirects-test-de-03", // Updated target
			RedirectionType: redirectstore.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
		{
			ID:              "2",
			ContentID:       "1",
			Source:          "/redirects-test-de-02",
			Target:          "/redirects-test-de-03", // New redirect added
			RedirectionType: redirectstore.RedirectionTypeAutomatic,
			Dimension:       "HMD-de",
		},
	}

	expectedDeleted := []redirectstore.EntityID{} // No deletions expected

	// Run the function
	updatedDefs, deletedIDs := rdutils.ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)

	// Assertions
	assert.Equal(t, len(expectedUpdated), len(updatedDefs), "Mismatch in updated redirect count")
	assert.Equal(t, len(expectedDeleted), len(deletedIDs), "Mismatch in deleted redirect count")

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
	// Existing redirects with a cycle: /a → /b → /c → /a
	oldRedirects := redirectstore.RedirectDefinitions{
		"/a": {ID: "1", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
		"/b": {ID: "2", ContentID: "2", Source: "/b", Target: "/c", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
		"/c": {ID: "3", ContentID: "3", Source: "/c", Target: "/a", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"}, // Cycle here
	}

	// New redirects (unchanged, but cycle still exists)
	newRedirects := []*redirectstore.RedirectDefinition{
		{ID: "4", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
		{ID: "5", ContentID: "2", Source: "/b", Target: "/c", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
		{ID: "6", ContentID: "3", Source: "/c", Target: "/a", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"}, // Cycle here
	}

	// No content nodes matter here
	currentNodes := map[string]*content.RepoNode{}

	// Run the function
	updatedDefs, deletedIDs := rdutils.ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)

	// 🔹 Debug: Print updated definitions
	t.Logf("Updated Redirects: %d", len(updatedDefs))
	for _, rd := range updatedDefs {
		t.Logf("Source: %s → Target: %s (Stale: %v)", rd.Source, rd.Target, rd.Stale)
	}

	// Assertions
	assert.Equal(t, len(newRedirects), len(updatedDefs), "Mismatch in updated redirect count")
	assert.Empty(t, deletedIDs, "No redirects should be deleted, only stale-marked")

	// Ensure cycle redirects are marked as stale
	for _, updated := range updatedDefs {
		if updated.Source == "/c" {
			assert.True(t, updated.Stale, "Redirect with cycle should be marked as stale")
		}
	}
}

func Test_ConsolidateRedirectDefinitions_NoCycle(t *testing.T) {
	// Existing redirects with a valid sequence: /a → /b → /c → /d
	oldRedirects := redirectstore.RedirectDefinitions{
		"/a": {ID: "1", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
		"/b": {ID: "2", ContentID: "2", Source: "/b", Target: "/c", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
		"/c": {ID: "3", ContentID: "3", Source: "/c", Target: "/d", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
	}

	// New redirects (updated target for `/b`)
	newRedirects := []*redirectstore.RedirectDefinition{
		{ID: "4", ContentID: "1", Source: "/a", Target: "/b", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
		{ID: "5", ContentID: "2", Source: "/b", Target: "/d", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"}, // Updated target
		{ID: "6", ContentID: "3", Source: "/c", Target: "/d", RedirectionType: redirectstore.RedirectionTypeAutomatic, Dimension: "global"},
	}

	// Available content nodes
	currentNodes := map[string]*content.RepoNode{
		"global": {ID: "1", URI: "/d"},
	}

	// Run the function
	updatedDefs, deletedIDs := rdutils.ConsolidateRedirectDefinitions(
		zap.L(),
		newRedirects,
		oldRedirects,
		currentNodes,
	)

	// Assertions
	assert.Equal(t, len(newRedirects), len(updatedDefs), "Mismatch in updated redirect count")
	assert.Empty(t, deletedIDs, "No redirects should be deleted")

	// Ensure the new target for `/b` is correctly updated
	for _, updated := range updatedDefs {
		if updated.Source == "/b" {
			assert.Equal(t, "/d", string(updated.Target), "Redirect target should be updated correctly")
		}
	}
}

// 🔹 Test Cases for HasCycle 🔹

func Test_HasCycle_DetectsCycle(t *testing.T) {
	redirects := map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/a"}, // Cycle: A → B → C → A
	}

	assert.True(t, rdutils.HasCycle("/a", "/b", redirects), "Cycle should be detected for /a → /b")
	assert.True(t, rdutils.HasCycle("/b", "/c", redirects), "Cycle should be detected for /b → /c")
	assert.True(t, rdutils.HasCycle("/c", "/a", redirects), "Cycle should be detected for /c → /a")
}

func Test_HasCycle_NoCycle(t *testing.T) {
	redirects := map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/d"}, // No cycle
	}

	assert.False(t, rdutils.HasCycle("/a", "/b", redirects), "No cycle should be detected for /a → /b")
	assert.False(t, rdutils.HasCycle("/b", "/c", redirects), "No cycle should be detected for /b → /c")
	assert.False(t, rdutils.HasCycle("/c", "/d", redirects), "No cycle should be detected for /c → /d")
}

func Test_HasCycle_SingleRedirect(t *testing.T) {
	redirects := map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
	}

	assert.False(t, rdutils.HasCycle("/a", "/b", redirects), "No cycle should be detected for /a → /b in isolation")
}

func Test_HasCycle_SelfReferencingRedirect(t *testing.T) {
	redirects := map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"/a": {Source: "/a", Target: "/a"}, // Self-referencing cycle
	}

	assert.True(t, rdutils.HasCycle("/a", "/a", redirects), "Self-referencing cycle should be detected for /a → /a")
}

func Test_HasCycle_LongChainNoCycle(t *testing.T) {
	redirects := map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/d"},
		"/d": {Source: "/d", Target: "/e"},
		"/e": {Source: "/e", Target: "/f"},
	}

	assert.False(t, rdutils.HasCycle("/a", "/b", redirects), "No cycle should be detected for a long chain without cycles")
}

func Test_HasCycle_ComplexCase(t *testing.T) {
	redirects := map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"/a": {Source: "/a", Target: "/b"},
		"/b": {Source: "/b", Target: "/c"},
		"/c": {Source: "/c", Target: "/d"},
		"/d": {Source: "/d", Target: "/b"}, // Cycle: B → C → D → B
	}

	assert.True(t, rdutils.HasCycle("/b", "/c", redirects), "Cycle should be detected for /b → /c")
	assert.True(t, rdutils.HasCycle("/c", "/d", redirects), "Cycle should be detected for /c → /d")
	assert.True(t, rdutils.HasCycle("/d", "/b", redirects), "Cycle should be detected for /d → /b")
}
