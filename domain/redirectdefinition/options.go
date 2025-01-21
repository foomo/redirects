package redirectdefinition

import (
	redirectprovider "github.com/foomo/redirects/pkg/provider"
)

func WithSiteIdentifierProvider(siteIdentifierFunc redirectprovider.SiteIdentifierProviderFunc) Option {
	return func(api *API) {
		api.getSiteIdentifierProvider = siteIdentifierFunc
	}
}

func WithRestrictedSourcesProvider(provider redirectprovider.RestrictedSourcesProviderFunc) Option {
	return func(api *API) {
		api.restrictedSourcesProvider = provider
	}
}
