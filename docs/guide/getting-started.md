# Getting started

`redirects` manages URL redirects for a site. It has two sides that never call
each other directly (see [Architecture](./architecture)):

- **Authoring** — the redirects service creates, validates and stores redirect
  definitions. Automatic redirects come from the content-server export; manual
  ones come from a frontend widget. Both are persisted in MongoDB and announced
  on NATS.
- **Serving** — the gateway keeps every active definition in memory, reloads on
  the NATS signal, and performs the actual HTTP redirect.

## Install

```sh
go get github.com/foomo/redirects/v2
```

## Prerequisites

| Dependency | Used by | Purpose |
| --- | --- | --- |
| MongoDB | redirects service | persists `RedirectDefinition` documents |
| NATS | both sides | change notification on the `redirects` topic |
| contentserver export | redirects service | source of automatic redirects |
| gateway | serving | pulls definitions and performs redirects |

## Minimal authoring setup

Wire the repository, the NATS update signal, the API and the service:

```go
import (
    cmrcmongo "github.com/bestbytes/commerce/pkg/persistence/mongo"
    "github.com/foomo/keel/log"
    keelservice "github.com/foomo/keel/service"
    "github.com/foomo/redirects/v2/domain/redirectdefinition"
    redirectrepository "github.com/foomo/redirects/v2/domain/redirectdefinition/repository"
    redirectservice "github.com/foomo/redirects/v2/domain/redirectdefinition/service"
    redirectnats "github.com/foomo/redirects/v2/pkg/nats"
)

persistor, err := cmrcmongo.New(ctx, "mongodb://localhost:27017/local")
log.Must(l, err, "failed to create persistor")

repo, err := redirectrepository.NewBaseRedirectsDefinitionRepository(l, persistor)
log.Must(l, err, "could not create redirect repository")

updateSignal, err := redirectnats.NewUpdateSignal(ctx, l,
    "nats://localhost:4222", "redirects", string(redirectnats.DefaultNatsTopic()))
log.Must(l, err, "failed to create update signal")

api, err := redirectdefinition.NewAPI(l, repo, updateSignal,
    redirectdefinition.WithSiteIdentifierProvider(siteIdentifier),
)
log.Must(l, err, "could not create redirect api")

service := redirectdefinition.NewService(l, api)

svr.AddServices(
    keelservice.NewHTTP(l, "redirects-admin", ":8080",
        redirectservice.NewDefaultAdminServiceGoTSRPCProxy(service)),
    keelservice.NewHTTP(l, "redirects-internal", ":8081",
        redirectservice.NewDefaultInternalServiceGoTSRPCProxy(service)),
)
```

That exposes two gotsrpc surfaces: the **admin** service for the widget and the
**internal** service for the content export and the gateway.

## Where to go next

- [Integration](./integration) — the full runbook: connecting the content-server
  export, wiring the gateway provider and middleware, and the NATS propagation
  flow, with a pre-flight checklist.
- [Architecture](./architecture) — how authoring and serving are decoupled, the
  write pipeline, and the request-matching algorithm.
- [Configuration reference](../reference/configuration) — every provider option,
  for both the API and the gateway provider.
- [API reference](../reference/api) — the admin and internal endpoints.
