package redirectprovider

import (
	"net/http"

	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
)

// BlacklistFunc reports whether a request must be left alone by the redirects provider.
// It is called in Process with the normalized request (trailing slash trimmed,
// query parameters sorted) before any redirect definition lookup.
type BlacklistFunc func(r *http.Request) bool

// defaultBlacklist is built once so DefaultBlacklist does not allocate per request.
var defaultBlacklist = AnyBlacklist(
	BlacklistHomepage,
	BlacklistPrefixes("/services", "/gateway"),
	BlacklistContains("/_next/"),
)

// DefaultBlacklist is the blacklist used when no WithBlacklistFunc option is given.
// It blocks the homepage, request URIs starting with /services or /gateway,
// and request URIs containing /_next/.
//
// Use the Blacklist* building blocks to assemble a variant of it, e.g. without /services:
//
//	AnyBlacklist(BlacklistHomepage, BlacklistPrefixes("/gateway"), BlacklistContains("/_next/"))
func DefaultBlacklist(r *http.Request) bool {
	return defaultBlacklist(r)
}

// AnyBlacklist combines blacklists: the request is blacklisted if any of them matches.
// Nil entries are ignored.
func AnyBlacklist(funcs ...BlacklistFunc) BlacklistFunc {
	return func(r *http.Request) bool {
		for _, f := range funcs {
			if f != nil && f(r) {
				return true
			}
		}

		return false
	}
}

// BlacklistHomepage blacklists requests to the homepage (path "/", regardless of query parameters).
func BlacklistHomepage(r *http.Request) bool {
	isHome, err := storex.RedirectRequest(r.URL.RequestURI()).IsHomepage()

	return err == nil && isHome
}

// BlacklistPrefixes blacklists requests whose request URI starts with one of the given prefixes.
// Matching is a plain string prefix match on the escaped request URI, so "/services" also matches "/services-foo".
func BlacklistPrefixes(prefixes ...string) BlacklistFunc {
	return func(r *http.Request) bool {
		return storex.RedirectRequest(r.URL.RequestURI()).HasPrefix(prefixes)
	}
}

// BlacklistContains blacklists requests whose request URI contains one of the given substrings.
func BlacklistContains(substrings ...string) BlacklistFunc {
	return func(r *http.Request) bool {
		return storex.RedirectRequest(r.URL.RequestURI()).Contains(substrings)
	}
}
