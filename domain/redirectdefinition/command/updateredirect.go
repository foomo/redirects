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
	// UpdateRedirect command
	UpdateRedirect struct {
		OldState map[string]*content.RepoNode `json:"oldState"`
		NewState map[string]*content.RepoNode `json:"newState"`
	}
	// UpdateRedirectHandlerFn handler
	UpdateRedirectHandlerFn func(ctx context.Context, l *zap.Logger, cmd UpdateRedirect) error
	// UpdateRedirectMiddlewareFn middleware
	UpdateRedirectMiddlewareFn func(next UpdateRedirectHandlerFn) UpdateRedirectHandlerFn
)

// UpdateRedirectHandler ...
func UpdateRedirectHandler(repo *redirectrepository.RedirectsDefinitionRepository) UpdateRedirectHandlerFn {
	return func(ctx context.Context, l *zap.Logger, cmd UpdateRedirect) error {
		return nil //repo.Upsert(ctx, entity)
	}
}

// UpdateRedirectHandlerComposed returns the handler with middleware applied to it
func UpdateRedirectHandlerComposed(handler UpdateRedirectHandlerFn, middlewares ...UpdateRedirectMiddlewareFn) UpdateRedirectHandlerFn {
	composed := func(next UpdateRedirectHandlerFn) UpdateRedirectHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger, cmd UpdateRedirect) error {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l, cmd)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger, cmd UpdateRedirect) error {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l, cmd)
	})
}
