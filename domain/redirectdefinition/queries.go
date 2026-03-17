package redirectdefinition

import (
	queryx "github.com/foomo/redirects/v2/domain/redirectdefinition/query"
)

type Queries struct {
	GetRedirects queryx.GetRedirectsHandlerFn
	Search       queryx.SearchHandlerFn
}
