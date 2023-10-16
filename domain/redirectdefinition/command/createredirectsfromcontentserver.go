package redirectcommand

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"github.com/foomo/contentserver/content"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectdefinitionutils "github.com/foomo/redirects/domain/redirectdefinition/utils"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

const dimension = "de"

type (
	// CreateRedirects command
	CreateRedirects struct {
		OldState map[string]*content.RepoNode `json:"oldState"`
		NewState map[string]*content.RepoNode `json:"newState"`
	}
	// CreateRedirectsHandlerFn handler
	CreateRedirectsHandlerFn func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error
	// CreateRedirectsMiddlewareFn middleware
	CreateRedirectsMiddlewareFn func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn
)

// CreateRedirectsHandler ...
func CreateRedirectsHandler(repo *redirectrepository.RedirectsDefinitionRepository) CreateRedirectsHandlerFn {
	return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
		newDefinitions, err := redirectdefinitionutils.AutoCreateRedirectDefinitions(l, cmd.OldState[dimension], cmd.NewState[dimension])
		if err != nil {
			return err
		}
		oldDefinitions, err := repo.FindAll(ctx)
		if err != nil {
			return err
		}
		consolidatedDefs, deletedDefs := redirectdefinitionutils.ConsolidateRedirectDefinitions(l, *oldDefinitions, newDefinitions)

		updateErr := repo.UpsertMany(ctx, &consolidatedDefs)
		if updateErr != nil {
			return updateErr
		}

		deleteErr := repo.DeleteMany(ctx, deletedDefs)
		if deleteErr != nil {
			return deleteErr
		}
		return nil
	}
}

// CreateRedirectsHandlerComposed returns the handler with middleware applied to it
func CreateRedirectsHandlerComposed(handler CreateRedirectsHandlerFn, middlewares ...CreateRedirectsMiddlewareFn) CreateRedirectsHandlerFn {
	composed := func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l, cmd)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l, cmd)
	})
}
