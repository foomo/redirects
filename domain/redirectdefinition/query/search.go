package redirectquery

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	redirectrepository "github.com/foomo/redirects/v2/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type (
	// Search query
	Search struct {
		Source       redirectstore.RedirectSource  `json:"source"`
		Dimension    redirectstore.Dimension       `json:"dimension"`
		ActiveState  redirectstore.ActiveStateType `json:"activeState"`
		Page         int                           `json:"page"`
		PageSize     int                           `json:"pageSize"`
		RedirectType redirectstore.RedirectionType `json:"type,omitempty"`
		Sort         redirectrepository.Sort       `json:"sort"`
	}
	// SearchHandlerFn handler
	SearchHandlerFn func(ctx context.Context, l *zap.Logger, qry Search) (*redirectstore.PaginatedResult, error)
	// SearchMiddlewareFn middleware
	SearchMiddlewareFn func(next SearchHandlerFn) SearchHandlerFn
)

// SearchHandler ...
func SearchHandler(repo redirectrepository.RedirectsDefinitionRepository) SearchHandlerFn {
	return func(ctx context.Context, _ *zap.Logger, qry Search) (*redirectstore.PaginatedResult, error) {
		// Default pagination values if not provided
		page := qry.Page
		if page < 1 {
			page = 1
		}
		pageSize := qry.PageSize
		if pageSize < 1 {
			pageSize = 20 // Default page size
		}

		// Validate RedirectType
		if !qry.RedirectType.IsValid() {
			return nil, fmt.Errorf("invalid redirect type: '%s'; should be empty, 'manual' or 'automatic'", qry.RedirectType)
		}

		// Validate ActiveState
		if !qry.ActiveState.IsValid() {
			return nil, fmt.Errorf("invalid active state: '%s'; should be empty, 'enabled' or 'disabled'", qry.RedirectType)
		}

		// Create pagination struct
		pagination := redirectrepository.Pagination{Page: page, PageSize: pageSize}

		return repo.FindMany(ctx, string(qry.Source), string(qry.Dimension), qry.RedirectType, qry.ActiveState, pagination, qry.Sort)
	}
}

// SearchHandlerComposed returns the handler with middleware applied to it
func SearchHandlerComposed(handler SearchHandlerFn, middlewares ...SearchMiddlewareFn) SearchHandlerFn {
	composed := func(next SearchHandlerFn) SearchHandlerFn {
		for _, middleware := range middlewares {
			localNext := next
			middlewareName := strings.Split(runtime.FuncForPC(reflect.ValueOf(middleware).Pointer()).Name(), ".")[2]
			next = middleware(func(ctx context.Context, l *zap.Logger, qry Search) (*redirectstore.PaginatedResult, error) {
				trace.SpanFromContext(ctx).AddEvent(middlewareName)
				return localNext(ctx, l, qry)
			})
		}
		return next
	}
	handlerName := strings.Split(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(), ".")[2]
	return composed(func(ctx context.Context, l *zap.Logger, qry Search) (*redirectstore.PaginatedResult, error) {
		trace.SpanFromContext(ctx).AddEvent(handlerName)
		return handler(ctx, l, qry)
	})
}
