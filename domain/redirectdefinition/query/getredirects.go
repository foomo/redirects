package redirectquery

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// GetRedirects query
	GetRedirects struct {
	}
	// GetRedirectsHandlerFn handler
	GetRedirectsHandlerFn func(ctx context.Context, l *zap.Logger) (*redirectstore.RedirectDefinitions, error)
	// GetRedirectsMiddlewareFn middleware
	GetRedirectsMiddlewareFn func(next GetRedirectsHandlerFn) GetRedirectsHandlerFn
)

// GetRedirectsHandler ...
func GetRedirectsHandler(repo *redirectrepository.RedirectsDefinitionRepository) GetRedirectsHandlerFn {
	return func(ctx context.Context, l *zap.Logger) (*redirectstore.RedirectDefinitions, error) {
		defs, err := repo.FindAll(ctx)
		if err != nil {
			return nil, err
		}
		return defs, nil
	}
}

// GetRedirectsHandlerComposed returns the handler with middleware applied to it
func GetRedirectsHandlerComposed(handler GetRedirectsHandlerFn, middlewares ...GetRedirectsMiddlewareFn) GetRedirectsHandlerFn {
	composed := func(next GetRedirectsHandlerFn) GetRedirectsHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger) (*redirectstore.RedirectDefinitions, error) {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger) (*redirectstore.RedirectDefinitions, error) {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l)
	})
}
