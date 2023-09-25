package redirectdefinition

import (
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

func ConsolidateRedirectDefinitions(
	l *zap.Logger,
	old, new *redirectstore.RedirectDefinitions,
) (*redirectstore.RedirectDefinitions, error) {
	return nil, nil
}
