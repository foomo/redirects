package redirectcommand

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	redirectnats "github.com/foomo/redirects/pkg/nats"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// UpdateRedirectsState command
	UpdateRedirectsState struct {
		RedirectDefinitions []*redirectstore.RedirectDefinition `json:"redirectDefinitions"`
	}
	// UpdateRedirectsStateHandlerFn handler
	UpdateRedirectsStateHandlerFn func(ctx context.Context, l *zap.Logger, cmd UpdateRedirectsState) error
	// UpdateRedirectsStateMiddlewareFn middleware
	UpdateRedirectsStateMiddlewareFn func(next UpdateRedirectsStateHandlerFn) UpdateRedirectsStateHandlerFn
)

// UpdateRedirectsStateHandler ...
func UpdateRedirectsStateHandler(repo redirectrepository.RedirectsDefinitionRepository) UpdateRedirectsStateHandlerFn {
	return func(ctx context.Context, _ *zap.Logger, cmd UpdateRedirectsState) error {
		return repo.UpsertMany(ctx, cmd.RedirectDefinitions)
	}
}

// UpdateRedirectsStateHandlerComposed returns the handler with middleware applied to it
func UpdateRedirectsStateHandlerComposed(handler UpdateRedirectsStateHandlerFn, middlewares ...UpdateRedirectsStateMiddlewareFn) UpdateRedirectsStateHandlerFn {
	composed := func(next UpdateRedirectsStateHandlerFn) UpdateRedirectsStateHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger, cmd UpdateRedirectsState) error {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l, cmd)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger, cmd UpdateRedirectsState) error {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l, cmd)
	})
}

// UpdateRedirectPublishMiddleware ...
func UpdateRedirectsStatePublishMiddleware(updateSignal *redirectnats.UpdateSignal) UpdateRedirectsStateMiddlewareFn {
	return func(next UpdateRedirectsStateHandlerFn) UpdateRedirectsStateHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd UpdateRedirectsState) error {
			err := next(ctx, l, cmd)
			if err != nil {
				return err
			}
			err = updateSignal.Publish()
			if err != nil {
				return err
			}
			return nil
		}
	}
}
