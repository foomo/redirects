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
		l.Info("calling create automatic redirects")
		newDefinitions, err := redirectdefinitionutils.AutoCreateRedirectDefinitions(l, cmd.OldState[dimension], cmd.NewState[dimension])
		if err != nil {
			l.Error("failed to execute auto create redirects", zap.Error(err))
			return err
		}
		oldDefinitions, err := repo.FindAll(ctx)
		if err != nil {
			l.Error("failed to fetch existing definitions", zap.Error(err))
			return err
		}
		l.Info("calling consolidate automatic redirects")
		consolidatedDefs, deletedDefs := redirectdefinitionutils.ConsolidateRedirectDefinitions(l, *oldDefinitions, newDefinitions)

		if len(consolidatedDefs) > 0 {
			updateErr := repo.UpsertMany(ctx, &consolidatedDefs)
			if updateErr != nil {
				l.Error("failed to updated definitions", zap.Error(updateErr))
				return updateErr
			}
		}
		if len(deletedDefs) > 0 {
			deleteErr := repo.DeleteMany(ctx, deletedDefs)
			if deleteErr != nil {
				l.Error("failed to delete definitions", zap.Error(deleteErr))
				return deleteErr
			}
		}

		l.Info("successfully finished create automatic redirects")
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
