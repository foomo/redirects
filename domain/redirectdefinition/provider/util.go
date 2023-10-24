package redirectprovider

import (
	"net/url"
	"path/filepath"
	"strings"

	store "github.com/foomo/redirects/domain/redirectdefinition/store"
)

func extractLegacyID(request store.RedirectRequest) (string, error) {
	url, err := url.Parse(string(request))
	if err != nil {
		return "", err
	}

	_, fileName := filepath.Split(url.Path)
	legacyID := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return legacyID, nil
}

func normalizeRedirectRequest(request store.RedirectRequest) (store.RedirectRequest, error) {
	url, err := url.Parse(string(request))
	if err != nil {
		return "", err
	}

	url.Path = strings.TrimSuffix(url.Path, "/")

	query, err := normalizeQueryString(url.RawQuery)
	if err != nil {
		return "", err
	}
	url.RawQuery = query

	return store.RedirectRequest(url.RequestURI()), nil
}

func normalizeQueryString(rawQueryString string) (string, error) {
	parsedQuery, err := url.ParseQuery(rawQueryString)
	if err != nil {
		return "", err
	}
	return parsedQuery.Encode(), nil
}

func mergeQueryStrings(base string, override string) (string, error) {
	baseValues, err := url.ParseQuery(base)
	if err != nil {
		return "", err
	}
	overrideValues, err := url.ParseQuery(override)
	if err != nil {
		return "", err
	}
	newValues := url.Values{}
	for k, v := range baseValues {
		newValues[k] = v
	}
	for k, v := range overrideValues {
		newValues[k] = v
	}
	query, err := normalizeQueryString(newValues.Encode())
	if err != nil {
		return "", err
	}
	return query, nil
}

func mergeQueryStringsFromURLs(base string, override string) (string, error) {
	urlBase, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	urlOverride, err := url.Parse(override)
	if err != nil {
		return "", err
	}
	query, err := mergeQueryStrings(urlBase.RawQuery, urlOverride.RawQuery)
	if err != nil {
		return "", err
	}
	urlOverride.RawQuery = query
	return urlOverride.RequestURI(), nil
}
