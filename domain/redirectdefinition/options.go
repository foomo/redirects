package redirectdefinition

import (
	"context"

	providerx "github.com/foomo/redirects/v2/pkg/provider"
)

// returns an empty list of restricted sources.
func defaultRestrictedSourcesProvider() []string {
	return []string{}
}

// returns "unknown" when user information is unavailable.
func defaultUserProvider(_ context.Context) string {
	return "unknown"
}

// returns false, meaning automatic redirects are enabled by default.
func defaultIsAutomaticRedirectInitiallyStaleProvider() bool {
	return false
}

func WithSiteIdentifierProvider(siteIdentifierFunc providerx.SiteIdentifierProviderFunc) Option {
	return func(api *API) {
		api.getSiteIdentifierProvider = siteIdentifierFunc
	}
}

func WithRestrictedSourcesProvider(provider providerx.RestrictedSourcesProviderFunc) Option {
	return func(api *API) {
		api.restrictedSourcesProvider = provider
	}
}

func WithUserProvider(provider providerx.UserProviderFunc) Option {
	return func(api *API) {
		api.userProvider = provider
	}
}

func WithIsAutomaticRedirectInitiallyStaleProvider(provider providerx.IsAutomaticRedirectInitiallyStaleProviderFunc) Option {
	return func(api *API) {
		api.isAutomaticRedirectInitiallyStaleProvider = provider
	}
}
