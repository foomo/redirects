package redirectcommand

import (
	"context"
	"fmt"
	"path"
	"strings"

	redirectrepository "github.com/foomo/redirects/domain/redirectdefinition/repository"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	redirectdefinitionutils "github.com/foomo/redirects/domain/redirectdefinition/utils"
	redirectprovider "github.com/foomo/redirects/pkg/provider"
	"go.uber.org/zap"
)

func validateRedirect(
	ctx context.Context,
	l *zap.Logger,
	repo redirectrepository.RedirectsDefinitionRepository,
	restrictedSourcesProvider redirectprovider.RestrictedSourcesProviderFunc,
	redirect *redirectstore.RedirectDefinition,
	next interface{},
) error {
	// Get restricted sources
	restrictedSources := []string{}
	if restrictedSourcesProvider != nil {
		restrictedSources = restrictedSourcesProvider()
	}

	// Convert source and target to lowercase
	source := strings.ToLower(string(redirect.Source))
	target := strings.ToLower(string(redirect.Target))

	if source == "/" {
		return fmt.Errorf("redirect from homepage is not allowed")
	}

	if source == target {
		return fmt.Errorf("redirect source and target cannot be the same")
	}

	for _, restricted := range restrictedSources {
		restricted = strings.ToLower(restricted)
		matched, _ := path.Match(restricted, source)
		if matched {
			return fmt.Errorf("source '%s' is restricted due to pattern '%s'", redirect.Source, restricted)
		}
	}

	// Fetch all existing redirects for the dimension
	existingRedirects, err := repo.FindAllByDimension(ctx, redirect.Dimension, false)
	if err != nil {
		return fmt.Errorf("failed to fetch existing redirects: %w", err)
	}

	// Check for cyclic redirect
	if redirectdefinitionutils.HasCycle(redirect.Source, redirect.Target, existingRedirects) {
		return fmt.Errorf("cyclic redirect detected: %s → %s creates a loop", redirect.Source, redirect.Target)
	}

	// Call the next handler dynamically based on function type
	switch fn := next.(type) {
	case CreateRedirectHandlerFn:
		return fn(ctx, l, CreateRedirect{RedirectDefinition: redirect})
	case UpdateRedirectHandlerFn:
		return fn(ctx, l, UpdateRedirect{RedirectDefinition: redirect})
	default:
		return fmt.Errorf("invalid handler type")
	}
}
