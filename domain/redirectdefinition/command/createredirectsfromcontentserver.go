package redirectcommand

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"github.com/foomo/contentserver/content"
	keellog "github.com/foomo/keel/log"
	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	redirectdefinitionutils "github.com/foomo/redirects/domain/redirectdefinition/utils"
	redirectnats "github.com/foomo/redirects/pkg/nats"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// CreateRedirects command
	CreateRedirects struct {
		OldState          map[string]*content.RepoNode        `json:"oldState"`
		NewState          map[string]*content.RepoNode        `json:"newState"`
		RedirectsToUpsert []*redirectstore.RedirectDefinition `json:"redirectsToUpsert,omitempty"`
		RedirectsToDelete []redirectstore.EntityID            `json:"redirectsToDeletee,omitempty"`
	}
	// CreateRedirectsHandlerFn handler
	CreateRedirectsHandlerFn func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error
	// CreateRedirectsMiddlewareFn middleware
	CreateRedirectsMiddlewareFn func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn
)

// CreateRedirectsHandler ...
func CreateRedirectsHandler(repo redirectrepository.RedirectsDefinitionRepository) CreateRedirectsHandlerFn {
	return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
		if len(cmd.RedirectsToUpsert) > 0 {
			updateErr := repo.UpsertMany(ctx, cmd.RedirectsToUpsert)
			if updateErr != nil {
				keellog.WithError(l, updateErr).Error("failed to updated definitions")
				return updateErr
			}
		}
		if len(cmd.RedirectsToDelete) > 0 {
			deleteErr := repo.DeleteMany(ctx, cmd.RedirectsToDelete)
			if deleteErr != nil {
				keellog.WithError(l, deleteErr).Error("failed to delete definitions")
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

// CreateRedirectsPublishMiddleware ...
func CreateRedirectsPublishMiddleware(updateSignal *redirectnats.UpdateSignal) CreateRedirectsMiddlewareFn {
	return func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
			err := next(ctx, l, cmd)
			if err != nil {
				return err
			}
			l.Info("publishing update signal")
			err = updateSignal.Publish()
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// CreateRedirectsAutoCreateMiddleware ...
func CreateRedirectsAutoCreateMiddleware() CreateRedirectsMiddlewareFn {
	return func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
			l.Info("auto creating redirects")
			dimensions := map[string]struct{}{}
			for dim := range cmd.OldState {
				dimensions[dim] = struct{}{}
			}

			for dim := range cmd.NewState {
				dimensions[dim] = struct{}{}
			}

			for dimension := range dimensions {
				oldNodeMap := redirectdefinitionutils.CreateFlatRepoNodeMap(cmd.OldState[dimension], make(map[string]*content.RepoNode))
				newNodeMap := redirectdefinitionutils.CreateFlatRepoNodeMap(cmd.NewState[dimension], make(map[string]*content.RepoNode))

				newDefinitions, err := redirectdefinitionutils.AutoCreateRedirectDefinitions(
					l,
					oldNodeMap,
					newNodeMap,
					redirectstore.Dimension(dimension),
				)
				if err != nil {
					keellog.WithError(l, err).Error("failed to execute auto create redirects")
					return err
				}
				cmd.RedirectsToUpsert = append(cmd.RedirectsToUpsert, newDefinitions...)
			}
			return next(ctx, l, cmd)
		}
	}
}

// CreateRedirectsConsolidateMiddleware ...
func CreateRedirectsConsolidateMiddleware(repo redirectrepository.RedirectsDefinitionRepository, autoDelete bool) CreateRedirectsMiddlewareFn {
	return func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
			l.Info("consolidating redirect definitions")
			redirectsToUpsert := []*redirectstore.RedirectDefinition{}
			redirectsToDelete := []redirectstore.EntityID{}

			// get all current definitions for the dimension from the database
			allCurrentDefinitions, err := repo.FindAll(ctx, true)
			if err != nil {
				l.Error("failed to fetch existing definitions", zap.Error(err))
				return err
			}
			for dimension, currentDefinitions := range allCurrentDefinitions {
				defs, ids := redirectdefinitionutils.ConsolidateRedirectDefinitions(
					l,
					cmd.RedirectsToUpsert,
					currentDefinitions,
					redirectdefinitionutils.CreateFlatRepoNodeMap(cmd.NewState[string(dimension)], make(map[string]*content.RepoNode)),
				)
				redirectsToUpsert = append(redirectsToUpsert, defs...)

				// if we are in auto delete mode we add the ids to the delete list
				// otherwise we soft delete the definitions
				if autoDelete {
					redirectsToDelete = append(redirectsToDelete, ids...)
				} else {
					softDeleteStrategy(ids, defs, currentDefinitions)
				}
			}
			cmd.RedirectsToUpsert = redirectsToUpsert
			cmd.RedirectsToDelete = redirectsToDelete

			return next(ctx, l, cmd)
		}
	}
}

// softDeleteStrategy ...
func softDeleteStrategy(
	idsToDelete []redirectstore.EntityID,
	newRedirects []*redirectstore.RedirectDefinition,
	currentDefinitions map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition,
) (redirectsToUpsert []*redirectstore.RedirectDefinition) {
	additionalRedirects := []*redirectstore.RedirectDefinition{}
	for _, id := range idsToDelete {
		for _, def := range newRedirects {
			if def.ID == id {
				def.Stale = true
				continue
			}
		}
		def, ok := currentDefinitions[id]
		if ok {
			def.Stale = true
			additionalRedirects = append(additionalRedirects, def)
		}
	}
	return append(newRedirects, additionalRedirects...)
}
