# API

The redirects service exposes two gotsrpc surfaces, generated from the interfaces
in `domain/redirectdefinition/service`:

| Service | Mount | Audience | Purpose |
| --- | --- | --- | --- |
| `AdminService` | `/services/redirectdefinition/admin` | frontend widget | manual CRUD + search |
| `InternalService` | `/services/redirectdefinition/internal` | content export, gateway | automatic creation + bulk read |

Mount them with the generated proxies:

```go
redirectservice.NewDefaultAdminServiceGoTSRPCProxy(service)
redirectservice.NewDefaultInternalServiceGoTSRPCProxy(service)
```

Consume them (e.g. from the gateway) with the generated clients:

```go
redirectservice.NewDefaultInternalServiceGoTSRPCClient("http://redirects-internal:8081")
redirectservice.NewDefaultAdminServiceGoTSRPCClient("http://redirects-admin:8080")
```

## AdminService

### `Search`

Search redirects for the resolved `{site}-{locale}` dimension, with pagination,
filtering and sorting. Defaults to page 1, page size 10.

```go
func (rs *Service) Search(
    _ http.ResponseWriter,
    r *http.Request,
    params *SearchParams,
) (*redirectstore.PaginatedResult, *redirectstore.RedirectDefinitionError)

type SearchParams struct {
    Locale       string
    Path         string
    Page         int
    PageSize     int
    RedirectType redirectstore.RedirectionType
    ActiveState  redirectstore.ActiveStateType
    Sort         redirectstore.Sort
}
```

### `Create`

Create a manual redirect. The `Dimension` is derived from the site provider and
the given `locale`; `LastUpdatedBy`/`Updated` are stamped from the user provider.
Validated for cycles and restricted sources.

```go
func (rs *Service) Create(
    _ http.ResponseWriter,
    r *http.Request,
    def *redirectstore.RedirectDefinition,
    locale string,
) (redirectstore.EntityID, *redirectstore.RedirectDefinitionError)
```

### `Update`

Update a redirect. Re-stamps `Updated`/`LastUpdatedBy` and re-validates.

```go
func (rs *Service) Update(
    _ http.ResponseWriter,
    r *http.Request,
    def *redirectstore.RedirectDefinition,
) *redirectstore.RedirectDefinitionError
```

### `Delete`

Delete a redirect by id.

```go
func (rs *Service) Delete(
    _ http.ResponseWriter,
    r *http.Request,
    id string,
) *redirectstore.RedirectDefinitionError
```

### `UpdateStates`

Enable or disable multiple redirects at once. `state = true` clears `Stale`
(enable); `state = false` sets it (disable).

```go
func (rs *Service) UpdateStates(
    _ http.ResponseWriter,
    r *http.Request,
    ids []*redirectstore.EntityID,
    state bool,
) *redirectstore.RedirectDefinitionError
```

## InternalService

### `CreateRedirectsFromContentserverexport`

Create automatic redirects from the content-server export diff, running cycle
detection before saving. No-op when [`WithEnabledFunc`](configuration#withenabledfunc)
returns `false`.

```go
func (rs *Service) CreateRedirectsFromContentserverexport(
    _ http.ResponseWriter,
    r *http.Request,
    oldState, newState map[string]*content.RepoNode,
) error
```

### `GetRedirects`

Return every stored definition, keyed by dimension then source. This is the
gateway provider's data source; its signature matches `RedirectsProviderFunc`.

```go
func (rs *Service) GetRedirects(
    _ http.ResponseWriter,
    r *http.Request,
) (map[redirectstore.Dimension]map[redirectstore.RedirectSource]*redirectstore.RedirectDefinition, error)
```

## Core types

```go
// the stored/served definition — see Architecture for field semantics
type RedirectDefinition struct {
    ID              EntityID
    ContentID       string
    Source          RedirectSource
    Target          RedirectTarget
    Code            RedirectCode
    RespectParams   bool
    TransferParams  bool
    RedirectionType RedirectionType
    Dimension       Dimension
    Stale           bool
    Updated         DateTime
    LastUpdatedBy   string
}

// the runtime redirect the gateway emits
type Redirect struct {
    Response RedirectResponse
    Code     RedirectCode
}

type PaginatedResult struct {
    Results  []*RedirectDefinition
    Total    int
    Page     int
    PageSize int
}
```

For the full, generated signatures see the
[godoc](https://pkg.go.dev/github.com/foomo/redirects/v2).
