# Configuration

Configuration is split across the two sides of the system: functional options on
the **authoring** `API`/`Service`, and options on the **serving**
`RedirectsProvider`.

## API options

Passed to `redirectdefinition.NewAPI(l, repo, updateSignal, opts...)`. Each
installs a provider the API calls at runtime. All are optional; defaults below.

### `WithSiteIdentifierProvider`

Resolves the `Site` for a request. The service combines it with the locale to
form the `Dimension` (`"{site}-{locale}"`) used by `Search` and `Create`.

```go
func WithSiteIdentifierProvider(f providerx.SiteIdentifierProviderFunc) Option

type SiteIdentifierProviderFunc func(r *http.Request) (store.Site, error)
```

**Default:** none — set this if you serve more than one site or use locales.

### `WithRestrictedSourcesProvider`

Supplies source paths that may **not** be used as redirect origins. Enforced when
validating manual create/update requests.

```go
func WithRestrictedSourcesProvider(p providerx.RestrictedSourcesProviderFunc) Option

type RestrictedSourcesProviderFunc func() []string
```

**Default:** an empty list — no sources are restricted.

### `WithUserProvider`

Reports the user responsible for a change, stored as `LastUpdatedBy` for audit
trails.

```go
func WithUserProvider(p providerx.UserProviderFunc) Option

type UserProviderFunc func(ctx context.Context) string
```

**Default:** returns `"unknown"`.

### `WithIsAutomaticRedirectInitiallyStaleProvider`

Decides whether newly created **automatic** redirects start `Stale` (inactive,
excluded from serving) or active.

```go
func WithIsAutomaticRedirectInitiallyStaleProvider(p providerx.IsAutomaticRedirectInitiallyStaleProviderFunc) Option

type IsAutomaticRedirectInitiallyStaleProviderFunc func() bool
```

**Default:** returns `false` — automatic redirects are active on creation.

## Service options

Passed to `redirectdefinition.NewService(l, api, opts...)`.

### `WithEnabledFunc`

Gates creation of automatic redirects. When it returns `false`,
`CreateRedirectsFromContentserverexport` returns early without creating anything —
letting you pause automatic redirects without a redeploy.

```go
func WithEnabledFunc(f func() bool) ServiceOption
```

**Default:** returns `true`.

## Gateway provider options

The serving side is built with
`redirectprovider.NewProvider(l, providerFunc, dimensionProviderFunc, updateChannel, opts...)`.

```go
func NewProvider(
    l *zap.Logger,
    providerFunc RedirectsProviderFunc,        // data source, e.g. InternalService client's GetRedirects
    dimensionProviderFunc DimensionProviderFunc, // resolves "{site}-{locale}" from the request
    updateChannel chan *nats.Msg,              // from NewUpdateSignalSubscribeChannel
    options ...RedirectsProviderOption,
) *RedirectsProvider

type RedirectsProviderFunc func(ctx context.Context) (map[store.Dimension]map[store.RedirectSource]*store.RedirectDefinition, error, error)
type DimensionProviderFunc func(r *http.Request) (store.Dimension, error)
```

### `WithMatcherFuncs`

Adds regex/custom matchers, tried after exact and path-only cache lookups fail.
Returns an error if no functions are supplied.

```go
func WithMatcherFuncs(matcherFuncs ...MatcherFunc) RedirectsProviderOption

type MatcherFunc func(r *http.Request) (*store.RedirectDefinition, error)
```

### `WithUseStandardRedirects`

Enables generic **standard redirects** — lower-casing the path and stripping a
trailing slash — as a `301` when no explicit definition matches.

```go
func WithUseStandardRedirects() RedirectsProviderOption
```

**Default:** off.

## NATS

Both sides connect with `redirectnats` helpers. The topic must match on both.

```go
const NatsTopicRedirects NatsTopic = "redirects" // DefaultNatsTopic()

// connection defaults (DefaultConnectOptions)
ConfigNatsTimeout       = 30 * time.Second // connect timeout
ConfigNatsReconnectWait = 1 * time.Second  // wait between reconnect attempts
```

Connections retry on failed connect by default. The service creates an
`UpdateSignal` with `NewUpdateSignal`; the gateway obtains a subscription channel
with `NewUpdateSignalSubscribeChannel`.

## Redirect codes

`RedirectDefinition.Code` accepts these values (`RedirectCode.Valid()`):

| Code | Constant | Meaning |
| --- | --- | --- |
| 301 | `RedirectCodePermanent` | Permanent redirect |
| 302 | `RedirectCodeFound` | Temporary redirect (found) |
| 307 | `RedirectCodeTemporary` | Temporary redirect, method preserved |
| 404 | `RedirectCodeNotFound` | Resource not found |
| 410 | `RedirectCodeGone` | Resource gone |
