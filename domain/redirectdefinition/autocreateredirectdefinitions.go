package redirectdefinition

import (
	"github.com/foomo/contentserver/content"
	redirectstore "github.com/foomo/redirects/domain/redirectdefinition/store"
	"go.uber.org/zap"
)

func AutoCreateRedirectDefinitions(
	l *zap.Logger,
	old, new map[string]*content.RepoNode,
) (redirectstore.RedirectDefinitions, error) {
	return nil, nil
}
