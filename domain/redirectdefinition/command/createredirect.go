package redirectcommand

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"github.com/foomo/contentserver/content"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// CreateRedirect command
	CreateRedirect struct {
		OldState map[string]*content.RepoNode `json:"oldState"`
		NewState map[string]*content.RepoNode `json:"newState"`
	}
	// CreateRedirectHandlerFn handler
	CreateRedirectHandlerFn func(ctx context.Context, l *zap.Logger, cmd CreateRedirect) error
	// CreateRedirectMiddlewareFn middleware
	CreateRedirectMiddlewareFn func(next CreateRedirectHandlerFn) CreateRedirectHandlerFn
)

// CreateRedirectHandler ...
func CreateRedirectHandler(repo *redirectrepository.RedirectsDefinitionRepository) CreateRedirectHandlerFn {
	return func(ctx context.Context, l *zap.Logger, cmd CreateRedirect) error {
		return nil //repo.Upsert(ctx, entity)
	}
}

// CreateRedirectHandlerComposed returns the handler with middleware applied to it
func CreateRedirectHandlerComposed(handler CreateRedirectHandlerFn, middlewares ...CreateRedirectMiddlewareFn) CreateRedirectHandlerFn {
	composed := func(next CreateRedirectHandlerFn) CreateRedirectHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger, cmd CreateRedirect) error {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l, cmd)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger, cmd CreateRedirect) error {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l, cmd)
	})
}
