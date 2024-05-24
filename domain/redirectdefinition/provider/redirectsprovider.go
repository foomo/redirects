package redirectprovider

import (
	"context"
	"regexp"
	"strings"

	keellog "github.com/foomo/keel/log"
	keelmongo "github.com/foomo/keel/persistence/mongo"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	store "github.com/foomo/redirects/domain/redirectdefinition/store"

	"go.uber.org/zap"
)

type ProductURLProvider func(ctx context.Context, legacyID string) (productURL string, clientErr error)
type MatcherFunc func(request store.RedirectRequest) (*store.RedirectDefinition, error)

type RedirectsProvider struct {
	l            *zap.Logger
	MatcherFuncs []MatcherFunc
	ctx          context.Context
	redirects    store.RedirectDefinitions
}

func NewRedirectsProvider(
	l *zap.Logger,
	ctx context.Context,
	regexLegacyProductURL *regexp.Regexp, //todo: get this from consumer because this is specific?
	persistor keelmongo.Persistor,
	useNats bool,
	matcherFuncs ...MatcherFunc) (*RedirectsProvider, error) {

	repo, errRepo := redirectrepository.NewBaseRedirectsDefinitionRepository(l, &persistor)
	if errRepo != nil {
		l.Error("failed to init redirects repository", zap.Error(errRepo))
		return nil, errRepo
	}

	redirects, err := repo.FindAll(ctx)
	if err != nil {
		l.Error("failed to find redirects", zap.Error(err))
		return nil, err
	}

	provider := &RedirectsProvider{
		l:         l,
		ctx:       ctx,
		redirects: *redirects,
	}

	// TODO: Do we need cache, nats and special func for loading
	//err := provider.LoadRedirects()
	//if err != nil {
	// error is already logged
	//return provider, err
	//}

	return provider, nil
}

func (p *RedirectsProvider) LoadRedirects() error {
	// TODO: is this needed
	//p.redirects =
	return nil
}

func (p *RedirectsProvider) Process(originalRequest store.RedirectRequest, dimension store.Dimension) (*store.Redirect, error) {
	p.l = keellog.With(p.l, zap.Any("originalRequest", originalRequest))
	p.l.Debug("process redirect request", zap.Any("original request", originalRequest))

	if p.isBlacklisted(originalRequest) {
		p.l.Debug("request is on black list", zap.Any("original request", originalRequest))
		return nil, nil
	}

	// normalize the incoming request url
	// a-z order of get-parameters
	request, err := normalizeRedirectRequest(originalRequest)
	if err != nil {
		keellog.WithError(p.l, err).Error("could normalized redirect request")
		return nil, err
	}

	redirect, ok := p.redirects[store.RedirectSource(request)][dimension]
	if ok && redirect != nil {
		return &store.Redirect{
			Response: store.RedirectResponse(redirect.Target),
			Code:     redirect.Code,
		}, nil
	}

	if !ok {
		definition, err := p.matchRedirectDefinition(request)

		if err != nil {
			keellog.WithError(p.l, err).Error("could not match redirect definition")
			return nil, err
		}

		// we found a redirect definition and process to create the response
		if definition != nil {
			redirect, err := p.createRedirect(request, definition)
			if err != nil {
				keellog.WithError(p.l, err).Error("could not create redirect response")
				return nil, err
			}
			return redirect, nil
		}
	}

	// if we do not find a specific redirect we check if we need to redirect
	// base on generic rules - no trailing slash/lowercased
	definition, err := p.checkForStandardRedirect(originalRequest)
	if err != nil {
		keellog.WithError(p.l, err).Error("could not check for standard redirect")
		return nil, err
	}
	if definition != nil {
		redirect, err := p.createRedirect(originalRequest, definition)
		if err != nil {
			keellog.WithError(p.l, err).Error("could not create redirect response")
			return nil, err
		}
		p.l.Debug("redirect based on standard rules", keellog.FValue(originalRequest))
		return redirect, nil
	}
	p.l.Debug("no redirect necessary", keellog.FValue(originalRequest))
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
func (p *RedirectsProvider) matchRedirectDefinition(request store.RedirectRequest) (*store.RedirectDefinition, error) {
	p.l.Debug("enter matchRedirectDefinition")

	// TODO: Do we need pool?
	// 1. full url from flat-pool
	// definition := p.matcherFull(request)
	// if definition != nil {
	// 	return definition, nil
	// }

	// 2. full url against regex
	definition, err := p.matcherFuncs(request)
	if err != nil {
		// no need to log anything here as logging is already done in .matcherFuncs
		return nil, err
	}
	if definition != nil {
		return definition, nil
	}

	// TODO: do wee need this
	// // 3. path and RespectParams from flat-pool
	// definition, err = p.matcherPath(request)
	// if err != nil {
	// 	// no need to log anything here as logging is already done in .matcherPath
	// 	return nil, err
	// }
	if definition != nil {
		return definition, nil
	}

	return nil, nil
}

// isBlacklisted define a series of paths/... redirection should leave alone
func (p *RedirectsProvider) isBlacklisted(request store.RedirectRequest) bool {
	p.l.Debug("enter isBlacklisted", zap.Any("request", request))
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

func (p *RedirectsProvider) matcherFuncs(request store.RedirectRequest) (*store.RedirectDefinition, error) {
	p.l.Debug("enter matcherLegacyProductURL")
	for _, matcherFunc := range p.MatcherFuncs {
		if redirectDefinition, err := matcherFunc(request); err != nil && redirectDefinition != nil {
			return redirectDefinition, nil
		}
	}
	return nil, nil
}

func (p *RedirectsProvider) createRedirect(request store.RedirectRequest, definition *store.RedirectDefinition) (*store.Redirect, error) {
	redirect := &store.Redirect{
		Code: definition.Code,
	}
	// if no transfer of parameters is allowed OR the request holds no query
	// the response is the definition's target
	if !definition.TransferParams || !strings.Contains(string(request), "?") {
		redirect.Response = store.RedirectResponse(definition.Target)
	} else {
		// merge query strings of the request and the target
		response, err := mergeQueryStringsFromURLs(string(request), string(definition.Target))
		if err != nil {
			keellog.WithError(p.l, err).Error("could not merge the query strings of the requests")
			return nil, err
		}
		redirect.Response = store.RedirectResponse(response)
	}
	return redirect, nil
}

func (p *RedirectsProvider) checkForStandardRedirect(redirectRequest store.RedirectRequest) (*store.RedirectDefinition, error) {
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
