package redirectcommand

import (
	"context"
	"fmt"
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
	// CreateRedirect command
	CreateRedirect struct {
		RedirectDefinition *redirectstore.RedirectDefinition `json:"redirectDefinition"`
	}
	// CreateRedirectHandlerFn handler
	CreateRedirectHandlerFn func(ctx context.Context, l *zap.Logger, cmd CreateRedirect) error
	// CreateRedirectMiddlewareFn middleware
	CreateRedirectMiddlewareFn func(next CreateRedirectHandlerFn) CreateRedirectHandlerFn
)

// CreateRedirectHandler ...
func CreateRedirectHandler(repo redirectrepository.RedirectsDefinitionRepository) CreateRedirectHandlerFn {
	return func(ctx context.Context, _ *zap.Logger, cmd CreateRedirect) error {
		return repo.Insert(ctx, cmd.RedirectDefinition)
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

// CreateRedirectPublishMiddleware ...
func CreateRedirectPublishMiddleware(updateSignal *redirectnats.UpdateSignal) CreateRedirectMiddlewareFn {
	return func(next CreateRedirectHandlerFn) CreateRedirectHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd CreateRedirect) error {
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

func ValidateRedirectMiddleware(restrictedPathsProvider redirectprovider.RestrictedPathsProvider) CreateRedirectMiddlewareFn {
	return func(next CreateRedirectHandlerFn) CreateRedirectHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd CreateRedirect) error {
			redirect := cmd.RedirectDefinition

			// Prevent '/' as source
			if redirect.Source == "/" {
				return fmt.Errorf("redirect source '/' is not allowed")
			}

			// Prevent same source and target
			if string(redirect.Source) == string(redirect.Target) {
				return fmt.Errorf("redirect source and target cannot be the same")
			}

			// Prevent restricted restrictedPaths as source
			restrictedPaths := restrictedPathsProvider()
			if restrictedPaths == nil {
				restrictedPaths = []string{}
			}

			for _, restricted := range restrictedPaths {
				if strings.HasPrefix(string(redirect.Source), restricted) {
					return fmt.Errorf("source '%s' is restricted", redirect.Source)
				}
			}

			// Proceed to next handler
			return next(ctx, l, cmd)
		}
	}
}
