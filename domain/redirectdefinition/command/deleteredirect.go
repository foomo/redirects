package redirectcommand

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
	// DeleteRedirect command
	DeleteRedirect struct {
		Source redirectstore.RedirectSource `json:"source"`
	}
	// DeleteRedirectHandlerFn handler
	DeleteRedirectHandlerFn func(ctx context.Context, l *zap.Logger, cmd DeleteRedirect) error
	// DeleteRedirectMiddlewareFn middleware
	DeleteRedirectMiddlewareFn func(next DeleteRedirectHandlerFn) DeleteRedirectHandlerFn
)

// DeleteRedirectHandler ...
func DeleteRedirectHandler(repo redirectrepository.BaseRedirectsDefinitionRepository) DeleteRedirectHandlerFn {
	return func(ctx context.Context, l *zap.Logger, cmd DeleteRedirect) error {
		return repo.Delete(ctx, string(cmd.Source))
	}
}

// DeleteRedirectHandlerComposed returns the handler with middleware applied to it
func DeleteRedirectHandlerComposed(handler DeleteRedirectHandlerFn, middlewares ...DeleteRedirectMiddlewareFn) DeleteRedirectHandlerFn {
	composed := func(next DeleteRedirectHandlerFn) DeleteRedirectHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger, cmd DeleteRedirect) error {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l, cmd)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger, cmd DeleteRedirect) error {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l, cmd)
	})
}
