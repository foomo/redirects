package redirectprovider

import (
	"net/http"
	"net/url"
)

func normalizeRedirectRequest(r *http.Request) (*http.Request, error) {
	request := r.Clone(r.Context())
	// request.URL.Path = strings.TrimSuffix(request.URL.Path, "/")
	query, err := normalizeQueryString(request.URL.RawQuery)
	if err != nil {
		return nil, err
	}
	request.URL.RawQuery = query
	return request, nil
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
