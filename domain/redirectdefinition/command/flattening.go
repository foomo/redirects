package redirectcommand

import (
	"context"

	repositoryx "github.com/foomo/redirects/v2/domain/redirectdefinition/repository"
	storex "github.com/foomo/redirects/v2/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

// applyFlattening retrieves all active redirects, applies flattening to resolve final targets,
// and updates only the changed redirects in the repository.
func applyFlattening(
	ctx context.Context,
	l *zap.Logger,
	repo repositoryx.RedirectsDefinitionRepository,
) error {
	// Fetch active redirects (non-stale)
	allRedirects, err := repo.FindAll(ctx, true)
	if err != nil {
		l.Error("Failed to fetch redirects for flattening", zap.Error(err))
		return err
	}

	flattenedRedirects := FlattenRedirects(allRedirects)

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

// FlattenRedirects applies flattening logic to active redirects
func FlattenRedirects(allRedirects map[storex.Dimension]map[storex.RedirectSource]*storex.RedirectDefinition) []*storex.RedirectDefinition {
	var flattened []*storex.RedirectDefinition

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
func resolveFinalTarget(target storex.RedirectTarget, redirects map[storex.RedirectSource]*storex.RedirectDefinition) storex.RedirectTarget {
	visited := make(map[string]struct{})

	for {
		nextRedirect, exists := redirects[storex.RedirectSource(target)]
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
