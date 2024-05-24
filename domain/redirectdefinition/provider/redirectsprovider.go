package redirectprovider

import (
	"context"
	"net/http"
	"strings"
	"sync"

	keellog "github.com/foomo/keel/log"
	store "github.com/foomo/redirects/domain/redirectdefinition/store"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type RedirectsProviderFunc func(ctx context.Context) (map[store.RedirectSource]*store.RedirectDefinition, error, error)
type MatcherFunc func(r *http.Request) (*store.RedirectDefinition, error)

type RedirectsProvider struct {
	sync.RWMutex
	l             *zap.Logger
	ctx           context.Context
	matcherFuncs  []MatcherFunc
	redirects     map[store.RedirectSource]*store.RedirectDefinition
	providerFunc  RedirectsProviderFunc
	updateChannel chan *nats.Msg
}

func NewProvider(
	l *zap.Logger,
	ctx context.Context,
	providerFunc RedirectsProviderFunc,
	updateChannel chan *nats.Msg,
	matcherFuncs ...MatcherFunc,
) *RedirectsProvider {
	provider := &RedirectsProvider{
		l:             l,
		ctx:           ctx,
		matcherFuncs:  matcherFuncs,
		providerFunc:  providerFunc,
		updateChannel: updateChannel,
	}
	return provider
}

func (p *RedirectsProvider) loadRedirects() error {
	redirectDefinitions, err, clientErr := p.providerFunc(p.ctx)
	if err != nil {
		return err
	} else if clientErr != nil {
		return clientErr
	} else if redirectDefinitions != nil {
		p.Lock()
		p.redirects = redirectDefinitions
		p.Unlock()
		return nil
	}
	return errors.New("no redirects loaded")
}

func (p *RedirectsProvider) Start() error {
	if err := p.loadRedirects(); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-p.updateChannel:
				if err := p.loadRedirects(); err != nil {
					keellog.WithError(p.l, err).Error("could not load redirects")
				}
			}
		}
	}()
	return nil
}

func (p *RedirectsProvider) Close(ctx context.Context) error {
	return nil
}

func (p *RedirectsProvider) Process(r *http.Request) (*store.Redirect, error) {
	l := keellog.With(p.l, keellog.FCodeMethod("Process"))
	l.Debug("process redirect request")

	// normalize the incoming request
	// a-z order of get-parameters
	request, err := normalizeRedirectRequest(r)
	if err != nil {
		keellog.WithError(l, err).Error("could normalized redirect request")
		return nil, err
	}

	// check if the request is on the blacklist
	if isBlacklisted(request) {
		l.Debug("request is on black list")
		return nil, nil
	}

	definition, err := p.matchRedirectDefinition(request)
	if err != nil {
		keellog.WithError(l, err).Error("could not match redirect definition")
		return nil, err
	}

	// we found a redirect definition and process to create the response
	if definition != nil {
		redirect, err := p.createRedirect(request, definition)
		if err != nil {
			keellog.WithError(l, err).Error("could not create redirect response")
			return nil, err
		}
		return redirect, nil
	}

	// if we do not find a specific redirect we check if we need to redirect
	// base on generic rules - no trailing slash/lowercased
	definition, err = p.checkForStandardRedirect(request)
	if err != nil {
		keellog.WithError(l, err).Error("could not check for standard redirect")
		return nil, err
	}
	if definition != nil {
		redirect, err := p.createRedirect(request, definition)
		if err != nil {
			keellog.WithError(l, err).Error("could not create redirect response")
			return nil, err
		}
		l.Debug("redirect based on standard rules")
		return redirect, nil
	}
	l.Debug("no redirect necessary")
	return nil, nil
}

// matchRedirectDefinition checks if there is a redirect definition matching the request
//
//	case 1: full string match
//	case 2:
//		request has query string
//		AND a definition matches the request path exactly
//		AND the definition allows a query via the RespectParams flag
//	case 3: request matches a source from the regex pool
func (p *RedirectsProvider) matchRedirectDefinition(r *http.Request) (*store.RedirectDefinition, error) {

	// 1. full url from cache
	p.RLock()
	definition, ok := p.redirects[store.RedirectSource(r.URL.RequestURI())]
	p.RUnlock()
	if ok {
		return definition, nil
	}

	// 2. full url against regex
	definition, err := p.execMatcherFuncs(r)
	if err != nil {
		// no need to log anything here as logging is already done in .matcherFuncs
		return nil, err
	} else if definition != nil {
		return definition, nil
	}

	if strings.Contains(r.URL.RequestURI(), "?") {
		p.RLock()
		definition, ok := p.redirects[store.RedirectSource(r.URL.Path)]
		p.RUnlock()
		if ok && definition.RespectParams {
			return definition, nil
		}
	}

	return nil, nil
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
func (p *RedirectsProvider) execMatcherFuncs(r *http.Request) (*store.RedirectDefinition, error) {
	for _, matcherFunc := range p.matcherFuncs {
		if definition, err := matcherFunc(r); err != nil && definition != nil {
			return definition, nil
		}
	}
	return nil, nil
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
func (p *RedirectsProvider) checkForStandardRedirect(r *http.Request) (*store.RedirectDefinition, error) {
	redirectRequest := store.RedirectRequest(r.URL.RequestURI())

	newRequest, redirectNeeded, err := redirectRequest.GenericTransform()
	if err != nil {
		return nil, err
	}
	if redirectNeeded {
		return &store.RedirectDefinition{
			ID:             "",
			Source:         store.RedirectSource(redirectRequest),
			Target:         store.RedirectTarget(newRequest),
			Code:           store.RedirectCodePermanent,
			RespectParams:  false,
			TransferParams: false,
		}, nil
	}
	return nil, nil
}
