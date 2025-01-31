package redirectcommand

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	redirectnats "github.com/foomo/redirects/pkg/nats"
	redirectprovider "github.com/foomo/redirects/pkg/provider"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// UpdateRedirect command
	UpdateRedirect struct {
		RedirectDefinition *redirectstore.RedirectDefinition `json:"redirectDefinition"`
	}
	// UpdateRedirectHandlerFn handler
	UpdateRedirectHandlerFn func(ctx context.Context, l *zap.Logger, cmd UpdateRedirect) error
	// UpdateRedirectMiddlewareFn middleware
	UpdateRedirectMiddlewareFn func(next UpdateRedirectHandlerFn) UpdateRedirectHandlerFn
)

// UpdateRedirectHandler ...
func UpdateRedirectHandler(repo redirectrepository.RedirectsDefinitionRepository) UpdateRedirectHandlerFn {
	return func(ctx context.Context, _ *zap.Logger, cmd UpdateRedirect) error {
		return repo.Update(ctx, cmd.RedirectDefinition)
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

// UpdateRedirectPublishMiddleware ...
func UpdateRedirectPublishMiddleware(updateSignal *redirectnats.UpdateSignal) UpdateRedirectMiddlewareFn {
	return func(next UpdateRedirectHandlerFn) UpdateRedirectHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd UpdateRedirect) error {
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

func ValidateUpdateRedirectMiddleware(restrictedSourcesProvider redirectprovider.RestrictedSourcesProviderFunc, repo redirectrepository.RedirectsDefinitionRepository) UpdateRedirectMiddlewareFn {
	return func(next UpdateRedirectHandlerFn) UpdateRedirectHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd UpdateRedirect) error {
			return validateRedirect(ctx, l, repo, restrictedSourcesProvider, cmd.RedirectDefinition, next)
		}
	}
}
