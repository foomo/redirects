package redirectquery

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	redirectrepository "github.com/foomo/redirects/v2/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// GetRedirects query
	GetRedirects struct {
	}
	// GetRedirectsHandlerFn handler
	GetRedirectsHandlerFn func(ctx context.Context, l *zap.Logger) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error)
	// GetRedirectsMiddlewareFn middleware
	GetRedirectsMiddlewareFn func(next GetRedirectsHandlerFn) GetRedirectsHandlerFn
)

// GetRedirectsHandler ...
func GetRedirectsHandler(repo redirectrepository.RedirectsDefinitionRepository) GetRedirectsHandlerFn {
	return func(ctx context.Context, _ *zap.Logger) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
		return repo.FindAll(ctx, true)
	}
}

// GetRedirectsHandlerComposed returns the handler with middleware applied to it
func GetRedirectsHandlerComposed(handler GetRedirectsHandlerFn, middlewares ...GetRedirectsMiddlewareFn) GetRedirectsHandlerFn {
	composed := func(next GetRedirectsHandlerFn) GetRedirectsHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l)
	})
}
