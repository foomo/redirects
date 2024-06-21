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
	// Search query
	Search struct {
		Source     redirectstore.RedirectSource `json:"source"`
		Dimension  redirectstore.Dimension      `json:"dimension"`
		OnlyActive bool                         `json:"onlyActive"`
	}
	// SearchHandlerFn handler
	SearchHandlerFn func(ctx context.Context, l *zap.Logger, qry Search) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error)
	// SearchMiddlewareFn middleware
	SearchMiddlewareFn func(next SearchHandlerFn) SearchHandlerFn
)

// SearchHandler ...
func SearchHandler(repo redirectrepository.RedirectsDefinitionRepository) SearchHandlerFn {
	return func(ctx context.Context, l *zap.Logger, qry Search) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
		return repo.FindMany(ctx, string(qry.Source), string(qry.Dimension), qry.OnlyActive)
	}
}

// SearchHandlerComposed returns the handler with middleware applied to it
func SearchHandlerComposed(handler SearchHandlerFn, middlewares ...SearchMiddlewareFn) SearchHandlerFn {
	composed := func(next SearchHandlerFn) SearchHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger, qry Search) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l, qry)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger, qry Search) (map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error) {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l, qry)
	})
}
