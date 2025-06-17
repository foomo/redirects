package redirectprovider

import (
	"context"
	"net/http"
	"strings"
	"sync"

	keellog "github.com/foomo/keel/log"
	store "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type RedirectsProviderInterface interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
	Process(r *http.Request) (*store.Redirect, error)
}
type DimensionProviderFunc func(r *http.Request) (store.Dimension, error)
type SiteIdentifierProviderFunc func(r *http.Request) (store.Site, error)
type RestrictedSourcesProviderFunc func() []string
type IsAutomaticRedirectInitiallyStaleProviderFunc func() bool
type UserProviderFunc func(ctx context.Context) string
type RedirectsProviderFunc func(ctx context.Context) (map[store.Dimension]map[store.RedirectSource]*store.RedirectDefinition, error, error)
type MatcherFunc func(r *http.Request) (*store.RedirectDefinition, error)

type RedirectsProviderOption func(provider *RedirectsProvider) error

type RedirectsProvider struct {
	sync.RWMutex
	l                     *zap.Logger
	redirects             map[store.Dimension]map[store.RedirectSource]*store.RedirectDefinition
	redirectsProviderFunc RedirectsProviderFunc
	dimensionProviderFunc DimensionProviderFunc
	updateChannel         chan *nats.Msg

	// optional features
	matcherFuncs         []MatcherFunc
	useStandardRedirects bool
}

func NewProvider(
	l *zap.Logger,
	providerFunc RedirectsProviderFunc,
	dimensionProviderFunc DimensionProviderFunc,
	updateChannel chan *nats.Msg,
	options ...RedirectsProviderOption,
) *RedirectsProvider {
	provider := &RedirectsProvider{
		l:                     l,
		redirectsProviderFunc: providerFunc,
		dimensionProviderFunc: dimensionProviderFunc,
		updateChannel:         updateChannel,
	}

	for _, opt := range options {
		if err := opt(provider); err != nil {
			keellog.WithError(l, err).Error("error applying provider option")
			continue
		}
	}

	return provider
}

func WithMatcherFuncs(matcherFuncs ...MatcherFunc) RedirectsProviderOption {
	return func(provider *RedirectsProvider) error {
		if len(matcherFuncs) == 0 {
			return errors.New("no matcher functions provided")
		}
		provider.matcherFuncs = matcherFuncs
		return nil
	}
}

func WithUseStandardRedirects() RedirectsProviderOption {
	return func(provider *RedirectsProvider) error {
		provider.useStandardRedirects = true
		return nil
	}
}

func (p *RedirectsProvider) loadRedirects(ctx context.Context) error {
	redirectDefinitions, err, clientErr := p.redirectsProviderFunc(ctx)
	if err != nil {
		return err
	}
	if clientErr != nil {
		return clientErr
	}
	if redirectDefinitions != nil {
		p.Lock()
		p.redirects = redirectDefinitions
		p.Unlock()
		return nil
	}
	return errors.New("no redirects loaded")
}

func (p *RedirectsProvider) Start(ctx context.Context) error {
	if err := p.loadRedirects(ctx); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.updateChannel:
				if err := p.loadRedirects(ctx); err != nil {
					keellog.WithError(p.l, err).Error("could not load redirects")
				}
			}
		}
	}()
	return nil
}

func (p *RedirectsProvider) Close(_ context.Context) error {
	return nil
}

func (p *RedirectsProvider) Process(r *http.Request) (redirect *store.Redirect, err error) {
	l := keellog.With(p.l, keellog.FCodeMethod("Process"))

	dimension, err := p.dimensionProviderFunc(r)
	if err != nil {
		return nil, err
	}

	// normalize the incoming request
	// a-z order of get-parameters
	request, err := normalizeRedirectRequest(r)
	if err != nil {
		keellog.WithError(l, err).Error("could not normalize redirect request")
		return nil, err
	}

	// check if the request is on the blacklist
	if isBlacklisted(request) {
		l.Debug("request is on black list")
		return redirect, nil
	}

	definition, err := p.matchRedirectDefinition(request, dimension)
	if err != nil {
		keellog.WithError(l, err).Error("could not match redirect definition")
		return nil, err
	}

	// we found a redirect definition and process to create the response
	if definition != nil {
		redirect, err = p.createRedirect(request, definition)
		if err != nil {
			keellog.WithError(l, err).Error("could not create redirect response")
			return nil, err
		}
		return redirect, nil
	}

	// if we do not find a specific redirect we check if we need to redirect
	// base on generic rules - no trailing slash/lowercased
	// only if enabled
	if p.useStandardRedirects {
		definition, err = p.checkForStandardRedirect(request)
		if err != nil {
			keellog.WithError(l, err).Error("could not check for standard redirect")
			return nil, err
		}
	}

	if definition == nil {
		l.Debug("no redirect necessary")
		return redirect, nil
	}

	redirect, err = p.createRedirect(request, definition)
	if err != nil {
		keellog.WithError(l, err).Error("could not create redirect response")
		return nil, err
	}

	l.Debug("redirect based on standard rules")
	return redirect, nil
}

// matchRedirectDefinition checks if there is a redirect definition matching the request
func (p *RedirectsProvider) matchRedirectDefinition(r *http.Request, dimension store.Dimension) (*store.RedirectDefinition, error) {
	// 1. full url from cache
	definition := p.definitionForDimensionAndSource(dimension, store.RedirectSource(r.URL.RequestURI()))
	if definition != nil {
		return definition, nil
	}

	if strings.Contains(r.URL.RequestURI(), "?") {
		definition := p.definitionForDimensionAndSource(dimension, store.RedirectSource(r.URL.Path))
		if definition != nil && definition.RespectParams {
			return definition, nil
		}
	}

	// 2. full url against regex
	definition, err := p.execMatcherFuncs(r)
	if err != nil {
		// no need to log anything here as logging is already done in .matcherFuncs
		return nil, err
	}

	return definition, nil
}

func (p *RedirectsProvider) definitionForDimensionAndSource(dimension store.Dimension, source store.RedirectSource) *store.RedirectDefinition {
	p.RLock()
	defer p.RUnlock()
	definitions, ok := p.redirects[dimension]
	if ok {
		definition, ok := definitions[source]
		if ok {
			return definition
		}
	}
	return nil
}

// isBlacklisted define a series of paths/... redirection should leave alone
func isBlacklisted(r *http.Request) bool {
	request := store.RedirectRequest(r.URL.RequestURI())

	isHome, err := request.IsHomepage()
	if err != nil {
		return false
	}
	prefixes := []string{"/services", "/gateway"}
	contains := []string{"/_next/"}
	if isHome || request.HasPrefix(prefixes) || request.Contains(contains) {
		return true
	}
	return false
}

// execMatcherFuncs executes the matcher functions
func (p *RedirectsProvider) execMatcherFuncs(r *http.Request) (definition *store.RedirectDefinition, err error) {
	for _, matcherFunc := range p.matcherFuncs {
		if definition, err = matcherFunc(r); err != nil && definition != nil {
			return definition, nil
		}
	}
	return definition, err
}

// createRedirect creates a redirect response based on the definition
func (p *RedirectsProvider) createRedirect(r *http.Request, definition *store.RedirectDefinition) (*store.Redirect, error) {
	redirect := &store.Redirect{
		Code: definition.Code,
	}
	// if no transfer of parameters is allowed OR the request holds no query
	// the response is the definition's target
	if !definition.TransferParams || !strings.Contains(r.URL.RequestURI(), "?") {
		redirect.Response = store.RedirectResponse(definition.Target)
	} else {
		// merge query strings of the request and the target
		response, err := mergeQueryStringsFromURLs(r.URL.RequestURI(), string(definition.Target))
		if err != nil {
			keellog.WithError(p.l, err).Error("could not merge the query strings of the requests")
			return nil, err
		}
		redirect.Response = store.RedirectResponse(response)
	}
	return redirect, nil
}

// checkForStandardRedirect checks if the request needs to be redirected based on generic rules
// it's possible to get no error and also to have no definition value, so the value needs to be checket after the method is called.
func (p *RedirectsProvider) checkForStandardRedirect(r *http.Request) (definition *store.RedirectDefinition, err error) {
	redirectRequest := store.RedirectRequest(r.URL.RequestURI())

	newRequest, redirectNeeded, err := redirectRequest.GenericTransform()
	if err != nil {
		return nil, err
	}
	if redirectNeeded {
		definition = &store.RedirectDefinition{
			ID:             "",
			Source:         store.RedirectSource(redirectRequest),
			Target:         store.RedirectTarget(newRequest),
			Code:           store.RedirectCodePermanent,
			RespectParams:  false,
			TransferParams: false,
		}
	}
	return definition, nil
}
