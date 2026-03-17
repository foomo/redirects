package redirectcommand

import (
	"context"
	"reflect"
	"runtime"
	"strings"

	"github.com/foomo/contentserver/content"
	keellog "github.com/foomo/keel/log"
	repositoryx "github.com/foomo/redirects/v2/domain/redirectdefinition/repository"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	utilsx "github.com/foomo/redirects/v2/domain/redirectdefinition/utils"
	natsx "github.com/foomo/redirects/v2/pkg/nats"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// CreateRedirects command
	CreateRedirects struct {
		OldState          map[string]*content.RepoNode `json:"oldState"`
		NewState          map[string]*content.RepoNode `json:"newState"`
		RedirectsToUpsert []*storex.RedirectDefinition `json:"redirectsToUpsert,omitempty"`
		RedirectsToDelete []storex.EntityID            `json:"redirectsToDeletee,omitempty"`
	}
	// CreateRedirectsHandlerFn handler
	CreateRedirectsHandlerFn func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error
	// CreateRedirectsMiddlewareFn middleware
	CreateRedirectsMiddlewareFn func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn
)

// CreateRedirectsHandler ...
func CreateRedirectsHandler(repo repositoryx.RedirectsDefinitionRepository) CreateRedirectsHandlerFn {
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
func CreateRedirectsPublishMiddleware(updateSignal *natsx.UpdateSignal, repo repositoryx.RedirectsDefinitionRepository) CreateRedirectsMiddlewareFn {
	return func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
			err := next(ctx, l, cmd)
			if err != nil {
				return err
			}

			if err := applyFlattening(ctx, l, repo); err != nil {
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
func CreateRedirectsAutoCreateMiddleware(initialStaleState bool) CreateRedirectsMiddlewareFn {
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
				oldNodeMap := utilsx.CreateFlatRepoNodeMap(cmd.OldState[dimension], make(map[string]*content.RepoNode))
				newNodeMap := utilsx.CreateFlatRepoNodeMap(cmd.NewState[dimension], make(map[string]*content.RepoNode))

				newDefinitions, err := utilsx.AutoCreateRedirectDefinitions(
					l,
					oldNodeMap,
					newNodeMap,
					storex.Dimension(dimension),
					initialStaleState,
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
func CreateRedirectsConsolidateMiddleware(repo repositoryx.RedirectsDefinitionRepository, autoDelete bool) CreateRedirectsMiddlewareFn {
	return func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn {
		return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
			l.Info("consolidating redirect definitions")

			redirectsToUpsert := []*storex.RedirectDefinition{}
			redirectsToDelete := []storex.EntityID{}

			// get all current definitions for the dimension from the database
			allCurrentDefinitions, err := repo.FindAll(ctx, true)
			if err != nil {
				l.Error("failed to fetch existing definitions", zap.Error(err))
				return err
			}

			for dimension, currentDefinitions := range allCurrentDefinitions {
				defs, ids := utilsx.ConsolidateRedirectDefinitions(
					l,
					cmd.RedirectsToUpsert,
					currentDefinitions,
					utilsx.CreateFlatRepoNodeMap(cmd.NewState[string(dimension)], make(map[string]*content.RepoNode)),
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
	idsToDelete []storex.EntityID,
	newRedirects []*storex.RedirectDefinition,
	currentDefinitions map[storex.RedirectSource]*storex.RedirectDefinition,
) []*storex.RedirectDefinition {
	additionalRedirects := []*storex.RedirectDefinition{}

	for _, id := range idsToDelete {
		for _, def := range newRedirects {
			if def.ID == id {
				def.Stale = true
				continue
			}
		}

		for _, def := range currentDefinitions {
			if def.ID == id {
				def.Stale = true
				additionalRedirects = append(additionalRedirects, def)
			}
		}
	}

	return append(newRedirects, additionalRedirects...)
}
