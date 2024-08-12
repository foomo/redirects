package redirectdefinition

import (
	redirectquery "github.com/foomo/redirects/domain/redirectdefinition/query"
)

type Queries struct {
	GetRedirects redirectquery.GetRedirectsHandlerFn
	Search       redirectquery.SearchHandlerFn
}
