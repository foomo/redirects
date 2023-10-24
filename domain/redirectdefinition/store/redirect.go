package redirectstore

import (
	"net/url"
	"strings"
)

const (
	RedirectCodePermanent RedirectCode = 301
	RedirectCodeTemporary RedirectCode = 307 // will this be needed?
)

type RedirectResponse string
type RedirectCode int
type Redirect struct {
	Response RedirectResponse
	Code     RedirectCode
}

func (r RedirectCode) Valid() bool {
	switch r {
	case
		RedirectCodePermanent:
		return true
	case
		RedirectCodeTemporary: // will this be needed
		return true
	default:
		return false
	}
}

// genericTransform checks whether the url need to be transformed to confirm to
// the url requirement - lowercased, no trailing slash
func (r RedirectRequest) GenericTransform() (newRequest RedirectRequest, hasChanged bool, err error) {
	newRequest = r
	base, err := url.Parse(string(r))
	if err != nil {
		return
	}
	base.Path = strings.TrimSuffix(strings.ToLower(base.Path), "/")
	newRequest = RedirectRequest(base.String())
	return newRequest, newRequest != r, nil
}

// isHomepage bool if we are on homepage
func (r RedirectRequest) IsHomepage() (isHome bool, err error) {
	base, err := url.Parse(string(r))
	if err != nil {
		return
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
