package redirectcommand

import (
	"context"

	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// applyFlattening retrieves all active redirects, applies flattening to resolve final targets,
// and updates only the changed redirects in the repository.
func applyFlattening(
	ctx context.Context,
	l *zap.Logger,
	repo redirectrepository.RedirectsDefinitionRepository,
) error {
	// Fetch active redirects (non-stale)
	allRedirects, err := repo.FindAll(ctx, true)
	if err != nil {
		l.Error("Failed to fetch redirects for flattening", zap.Error(err))
		return err
	}

	flattenedRedirects := flattenRedirects(allRedirects)

	// If no redirects changed, avoid unnecessary DB writes
	if len(flattenedRedirects) == 0 {
		l.Info("No redirects changed after flattening")
		return nil
	}

	// Persist only changed redirects
	if err := repo.UpsertMany(ctx, flattenedRedirects); err != nil {
		l.Error("Failed to persist flattened redirects", zap.Error(err))
		return err
	}

	l.Info("Successfully updated changed redirects", zap.Int("count", len(flattenedRedirects)))
	return nil
}

// flattenRedirects applies flattening logic to active redirects
func flattenRedirects(allRedirects map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition) []*redirectstore.RedirectDefinition {
	var flattened []*redirectstore.RedirectDefinition

	for _, redirectsBySource := range allRedirects {
		for _, redirect := range redirectsBySource {
			// Resolve final target by flattening the chain
			finalTarget := resolveFinalTarget(redirect.Target, redirectsBySource)

			// Only store changes (avoid unnecessary updates)
			if finalTarget != redirect.Target {
				redirect.Target = finalTarget
				flattened = append(flattened, redirect)
			}
		}
	}

	return flattened
}

// resolveFinalTarget follows the redirect chain to find the final target
func resolveFinalTarget(target redirectstore.RedirectTarget, redirects map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition) redirectstore.RedirectTarget {
	visited := make(map[string]struct{})

	for {
		nextRedirect, exists := redirects[redirectstore.RedirectSource(target)]
		if !exists || nextRedirect.Target == "" {
			return target
		}

		// Prevent infinite loops
		if _, seen := visited[string(target)]; seen {
			return target
		}

		visited[string(target)] = struct{}{}
		target = nextRedirect.Target
	}
}
