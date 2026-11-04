package redirectstore

import (
	"net/url"
	"strings"
)

const (
	RedirectCodePermanent RedirectCode = 301 // Permanent Redirect
	RedirectCodeFound     RedirectCode = 302 // Temporary Redirect (Found)
	RedirectCodeTemporary RedirectCode = 307 // Temporary Redirect with method preservation
	RedirectCodeNotFound  RedirectCode = 404 // Resource Not Found
	RedirectCodeGone      RedirectCode = 410 // Resource Gone
)

type RedirectResponse string
type RedirectCode int
type Redirect struct {
	Response RedirectResponse
	Code     RedirectCode
}

func (r RedirectCode) Valid() bool {
	switch r {
	case RedirectCodePermanent,
		RedirectCodeFound,
		RedirectCodeTemporary,
		RedirectCodeNotFound,
		RedirectCodeGone:
		return true
	default:
		return false
	}
}

// GenericTransform checks whether the url need to be transformed to confirm to
// the url requirement - lowercased, no trailing slash.
// Only path-only request URIs are transformed: anything that parses to a scheme
// or host (e.g. "//evil.com/Foo", "/%2Fevil.com/Foo") or that browsers would
// treat as protocol-relative ("/\evil.com") is never redirected, to prevent open redirects.
func (r RedirectRequest) GenericTransform() (RedirectRequest, bool, error) {
	base, err := url.Parse(string(r))
	if err != nil {
		return r, false, err
	}

	if base.Scheme != "" || base.Host != "" || strings.HasPrefix(base.Path, "//") || strings.HasPrefix(base.Path, "/\\") {
		return r, false, nil
	}

	base.Path = strings.TrimSuffix(strings.ToLower(base.Path), "/")
	newRequest := RedirectRequest(base.String())

	return newRequest, newRequest != r, nil
}

// IsHomepage returns true if we are on homepage
func (r RedirectRequest) IsHomepage() (bool, error) {
	base, err := url.Parse(string(r))
	if err != nil {
		return false, err
	}

	return base.Path == "/", nil
}

func (r RedirectRequest) HasPrefix(patterns []string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(string(r), pattern) {
			return true
		}
	}

	return false
}

func (r RedirectRequest) Contains(patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(string(r), pattern) {
			return true
		}
	}

	return false
}

func (r RedirectRequest) String() string {
	return string(r)
}
