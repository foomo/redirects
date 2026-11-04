package redirectprovider_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	redirectprovider "github.com/foomo/redirects/v2/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const testDimension storex.Dimension = "de"

func newTestProvider(t *testing.T, definitions []*storex.RedirectDefinition, opts ...redirectprovider.RedirectsProviderOption) *redirectprovider.RedirectsProvider {
	t.Helper()

	redirects := map[storex.RedirectSource]*storex.RedirectDefinition{}
	for _, d := range definitions {
		redirects[d.Source] = d
	}

	provider := redirectprovider.NewProvider(
		zap.NewNop(),
		func(_ context.Context) (map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition, error, error) {
			return map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition{testDimension: redirects}, nil, nil
		},
		func(_ *http.Request) (storex.Dimension, error) { return testDimension, nil },
		nil,
		opts...,
	)
	require.NoError(t, provider.Start(t.Context()))

	return provider
}

func newRequest(t *testing.T, uri string) *http.Request {
	t.Helper()

	return httptest.NewRequestWithContext(t.Context(), http.MethodGet, uri, nil)
}

func definition(source, target string) *storex.RedirectDefinition {
	return &storex.RedirectDefinition{
		Source: storex.RedirectSource(source),
		Target: storex.RedirectTarget(target),
		Code:   storex.RedirectCodePermanent,
	}
}

// bksBlacklist is DefaultBlacklist without /services, as a consumer would compose it.
var bksBlacklist = redirectprovider.AnyBlacklist(
	redirectprovider.BlacklistHomepage,
	redirectprovider.BlacklistPrefixes("/gateway"),
	redirectprovider.BlacklistContains("/_next/"),
)

func TestDefaultBlacklist(t *testing.T) {
	t.Parallel()

	tests := []struct {
		uri  string
		want bool
	}{
		{"/", true},
		{"/?foo=bar", true},
		{"/services", true},
		{"/services/internetbanking-und-apps/bks-app-2.0", true},
		{"/servicesfoo", true}, // plain prefix match, no segment boundary
		{"/gateway/foo", true},
		{"/_next/static/chunk.js", true},
		{"/de/_next/data/page.json", true},
		{"/foo", false},
		{"/de/services", false},
		{"/foo?q=/services", false},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, redirectprovider.DefaultBlacklist(newRequest(t, tt.uri)))
		})
	}
}

func TestBlacklistBuildingBlocks(t *testing.T) {
	t.Parallel()

	assert.True(t, bksBlacklist(newRequest(t, "/")))
	assert.True(t, bksBlacklist(newRequest(t, "/gateway/foo")))
	assert.True(t, bksBlacklist(newRequest(t, "/_next/static/chunk.js")))
	assert.False(t, bksBlacklist(newRequest(t, "/services/internetbanking-und-apps/bks-app-2.0")))

	assert.False(t, redirectprovider.AnyBlacklist()(newRequest(t, "/")))
	assert.False(t, redirectprovider.AnyBlacklist(nil)(newRequest(t, "/")))
}

func TestProcess_DefaultBlacklist(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, []*storex.RedirectDefinition{
		definition("/services/internetbanking-und-apps/bks-app-2.0", "/services/internetbanking-und-apps/bks-app"),
		definition("/gateway/foo", "/bar"),
		definition("/_next/foo", "/bar"),
		definition("/foo", "/bar"),
	})

	for _, uri := range []string{
		"/",
		"/services/internetbanking-und-apps/bks-app-2.0",
		"/gateway/foo",
		"/_next/foo",
	} {
		redirect, err := provider.Process(newRequest(t, uri))
		require.NoError(t, err, uri)
		assert.Nil(t, redirect, uri)
	}

	// sanity check: a non-blacklisted source is redirected
	redirect, err := provider.Process(newRequest(t, "/foo"))
	require.NoError(t, err)
	require.NotNil(t, redirect)
	assert.Equal(t, storex.RedirectResponse("/bar"), redirect.Response)
}

func TestProcess_WithBlacklistFunc(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t,
		[]*storex.RedirectDefinition{
			definition("/services/internetbanking-und-apps/bks-app-2.0", "/services/internetbanking-und-apps/bks-app"),
			definition("/gateway/foo", "/bar"),
		},
		redirectprovider.WithBlacklistFunc(bksBlacklist),
	)

	redirect, err := provider.Process(newRequest(t, "/services/internetbanking-und-apps/bks-app-2.0"))
	require.NoError(t, err)
	require.NotNil(t, redirect)
	assert.Equal(t, storex.RedirectResponse("/services/internetbanking-und-apps/bks-app"), redirect.Response)
	assert.Equal(t, storex.RedirectCodePermanent, redirect.Code)

	// remaining entries of the custom blacklist still apply
	redirect, err = provider.Process(newRequest(t, "/gateway/foo"))
	require.NoError(t, err)
	assert.Nil(t, redirect)
}

func TestProcess_WithBlacklistFuncNil(t *testing.T) {
	t.Parallel()

	// the option errors, NewProvider logs and skips it, so the default stays in place
	provider := newTestProvider(t,
		[]*storex.RedirectDefinition{definition("/services/foo", "/bar")},
		redirectprovider.WithBlacklistFunc(nil),
	)

	redirect, err := provider.Process(newRequest(t, "/services/foo"))
	require.NoError(t, err)
	assert.Nil(t, redirect)
}

func TestProcess_BlacklistRunsBeforeLookup(t *testing.T) {
	t.Parallel()

	var (
		blacklistCalls int
		matcherCalls   int
		seenURI        string
	)

	provider := newTestProvider(t,
		// exact match for the normalized request URI
		[]*storex.RedirectDefinition{definition("/Blocked?a=1&b=2", "/target")},
		redirectprovider.WithBlacklistFunc(func(r *http.Request) bool {
			blacklistCalls++
			seenURI = r.URL.RequestURI()

			return strings.HasPrefix(strings.ToLower(r.URL.Path), "/blocked")
		}),
		redirectprovider.WithMatcherFuncs(func(_ *http.Request) (*storex.RedirectDefinition, error) {
			matcherCalls++
			return definition("/blocked", "/from-matcher"), nil
		}),
		// "/Blocked" would otherwise also be a standard redirect (lowercase)
		redirectprovider.WithUseStandardRedirects(),
	)

	redirect, err := provider.Process(newRequest(t, "/Blocked/?b=2&a=1"))
	require.NoError(t, err)
	assert.Nil(t, redirect, "exact definition must not be used")

	// the blacklist receives the normalized request
	assert.Equal(t, 1, blacklistCalls)
	assert.Equal(t, "/Blocked?a=1&b=2", seenURI)

	// no exact definition: the matcher funcs would be next, but must not run
	redirect, err = provider.Process(newRequest(t, "/blocked/matcher-only"))
	require.NoError(t, err)
	assert.Nil(t, redirect)
	assert.Equal(t, 2, blacklistCalls)
	assert.Zero(t, matcherCalls, "matcher funcs must not run for blacklisted requests")
}

func TestProcess_StandardRedirectNoOpenRedirect(t *testing.T) {
	t.Parallel()

	provider := newTestProvider(t, nil, redirectprovider.WithUseStandardRedirects())

	for _, uri := range []string{"//evil.com/Foo/", "/%2Fevil.com/Foo", "/\\evil.com/Foo/"} {
		req := newRequest(t, "/")
		// set the raw request target like net/http's server would for "GET <uri>"
		u, err := url.ParseRequestURI(uri)
		require.NoError(t, err)

		req.URL = u
		req.RequestURI = uri

		redirect, err := provider.Process(req)
		require.NoError(t, err, uri)
		assert.Nil(t, redirect, uri)
	}

	// regular standard redirects keep working
	redirect, err := provider.Process(newRequest(t, "/Foo/"))
	require.NoError(t, err)
	require.NotNil(t, redirect)
	assert.Equal(t, storex.RedirectResponse("/foo"), redirect.Response)
}
