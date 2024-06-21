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
		OldState map[string]*content.RepoNode `json:"oldState"`
		NewState map[string]*content.RepoNode `json:"newState"`
	}
	// CreateRedirectsHandlerFn handler
	CreateRedirectsHandlerFn func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error
	// CreateRedirectsMiddlewareFn middleware
	CreateRedirectsMiddlewareFn func(next CreateRedirectsHandlerFn) CreateRedirectsHandlerFn
)

// CreateRedirectsHandler ...
func CreateRedirectsHandler(repo redirectrepository.RedirectsDefinitionRepository) CreateRedirectsHandlerFn {
	return func(ctx context.Context, l *zap.Logger, cmd CreateRedirects) error {
		l.Info("calling create automatic redirects")

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

			// get all current definitions for the dimension from the database
			currentDefinitions, err := repo.FindMany(ctx, "", dimension)
			if err != nil {
				l.Error("failed to fetch existing definitions", zap.Error(err))
				return err
			}

			consolidatedDefs, deletedIDs := redirectdefinitionutils.ConsolidateRedirectDefinitions(
				l,
				newDefinitions,
				currentDefinitions,
				newNodeMap,
			)

			l.Info("consolidated definitions", zap.Any("consolidatedDefs", consolidatedDefs), zap.Any("deletedIDs", deletedIDs))

			if len(consolidatedDefs) > 0 {
				updateErr := repo.UpsertMany(ctx, &consolidatedDefs)
				if updateErr != nil {
					keellog.WithError(l, updateErr).Error("failed to updated definitions")
					return updateErr
				}
			}
			if len(deletedIDs) > 0 {
				deleteErr := repo.DeleteMany(ctx, deletedIDs)
				if deleteErr != nil {
					keellog.WithError(l, deleteErr).Error("failed to delete definitions")
					return deleteErr
				}
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
			err = updateSignal.Publish()
			if err != nil {
				return err
			}
			return nil
		}
	}
}
