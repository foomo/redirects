package redirectdefinition

import (
	redirectprovider "github.com/foomo/redirects/pkg/provider"
)

func WithSiteIdentifierProvider(siteIdentifierFunc redirectprovider.SiteIdentifierProviderFunc) Option {
	return func(api *API) {
		api.getSiteIdentifierProvider = siteIdentifierFunc
	}
}

func WithRestrictedPathsProvider(provider redirectprovider.RestrictedPathsProviderFunc) Option {
	return func(api *API) {
		api.restrictedPathsProvider = provider
	}
}
