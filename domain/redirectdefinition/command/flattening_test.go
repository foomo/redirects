package redirectcommand_test

import (
	"testing"

	redirectcommand "github.com/foomo/redirects/domain/redirectdefinition/command"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"github.com/stretchr/testify/assert"
)

func Test_FlattenRedirects_SimpleChain(t *testing.T) {
	redirects := map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"global": {
			"/a": {Source: "/a", Target: "/b", RedirectionType: redirectstore.RedirectionTypeAutomatic, Stale: false},
			"/b": {Source: "/b", Target: "/c", RedirectionType: redirectstore.RedirectionTypeManual, Stale: false},
			"/c": {Source: "/c", Target: "/final", RedirectionType: redirectstore.RedirectionTypeAutomatic, Stale: false},
		},
	}
	flattened := redirectcommand.FlattenRedirects(redirects)

	// Assertions
	assert.Len(t, flattened, 3)
	assert.Equal(t, "/final", string(flattened[0].Target))
	assert.Equal(t, "/final", string(flattened[1].Target))
	assert.Equal(t, "/final", string(flattened[2].Target))
}

func Test_FlattenRedirects_TwoFlatten(t *testing.T) {
	redirects := map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"global": {
			"/a": {Source: "/a", Target: "/b", RedirectionType: redirectstore.RedirectionTypeAutomatic, Stale: false},
			"/b": {Source: "/b", Target: "/f", RedirectionType: redirectstore.RedirectionTypeManual, Stale: false},
			"/c": {Source: "/c", Target: "/e", RedirectionType: redirectstore.RedirectionTypeAutomatic, Stale: false},
		},
	}

	flattened := redirectcommand.FlattenRedirects(redirects)

	// Assertions
	assert.Len(t, flattened, 3)
	assert.Equal(t, "/f", string(flattened[0].Target))
	assert.Equal(t, "/f", string(flattened[1].Target))
	assert.Equal(t, "/e", string(flattened[2].Target))
}

func Test_FlattenRedirects_MultipleSourcesToSameTarget(t *testing.T) {
	redirects := map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition{
		"global": {
			"/a": {Source: "/a", Target: "/c", RedirectionType: redirectstore.RedirectionTypeAutomatic, Stale: false},
			"/b": {Source: "/b", Target: "/c", RedirectionType: redirectstore.RedirectionTypeManual, Stale: false},
			"/c": {Source: "/c", Target: "/final", RedirectionType: redirectstore.RedirectionTypeAutomatic, Stale: false},
		},
	}

	flattened := redirectcommand.FlattenRedirects(redirects)

	// Assertions: Both /a and /b should point to /final
	assert.Len(t, flattened, 3)
	assert.Equal(t, "/final", string(flattened[0].Target), "/a should flatten to /final")
	assert.Equal(t, "/final", string(flattened[1].Target), "/b should flatten to /final")
	assert.Equal(t, "/final", string(flattened[2].Target), "/c should flatten to /final")
}
