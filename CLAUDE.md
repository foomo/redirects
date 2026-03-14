# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`github.com/foomo/redirects/v2` is a Go library (no binaries) for managing URL redirects with automatic generation from content changes, cycle detection, and chain flattening. Uses MongoDB for storage and NATS for signaling. Uses `github.com/stretchr/testify` for testing.

## Common Commands

```bash
make test          # Run tests (outputs coverage.out)
make test.race     # Run tests with race detector
make test.cover    # Tests with coverage report (HTML)
make lint          # Run golangci-lint
make lint.fix      # Run golangci-lint with --fix
make fmt           # Format code (golangci-lint fmt)
make tidy          # go mod tidy
make generate      # go generate ./...
```

Run a single test: `go test -run TestName ./domain/redirectdefinition/command/`

Prerequisites: `mise` (tool manager) and `lefthook` (git hooks) must be installed. Run `mise install` to set up tool versions.

## Linting

- golangci-lint v2 config with `default: all` (all linters enabled, specific ones disabled)
- Formatters: `gofmt` and `goimports`
- `importas` enforces that internal `foomo/redirects` imports use an alias with `x` suffix (e.g., `redirectstorex`)
- `testpackage` linter requires tests in separate `_test` packages
- Pre-commit hook auto-formats staged `.go` files via `golangci-lint fmt`

## CI Pipeline

CI runs on push to `main` and PRs:
1. `make tidy` + `make generate` + `make fmt` — must produce no diffs
2. `make lint`
3. `make test`

## Architecture

Domain-Driven Design with a single domain: `redirectdefinition`.

- **`domain/redirectdefinition/api.go`** — API orchestrator, wires commands and queries
- **`domain/redirectdefinition/service.go`** — HTTP service layer (AdminService + InternalService via GoTSRPC)
- **`domain/redirectdefinition/command/`** — Write operations (create, update, delete, flattening, validation). Each command uses a middleware chain for validation and publishing
- **`domain/redirectdefinition/query/`** — Read operations (get redirects, paginated search)
- **`domain/redirectdefinition/repository/`** — MongoDB repository with `foomo/keel` persistence layer
- **`domain/redirectdefinition/store/`** — Domain models and value objects (RedirectDefinition, EntityID, filters, sorting)
- **`domain/redirectdefinition/utils/`** — Consolidation and auto-creation logic
- **`pkg/middleware/`** — HTTP middleware for applying redirects
- **`pkg/nats/`** — NATS update signal integration
- **`pkg/provider/`** — Redirect data provider for consumers

Key provider functions configure behavior: `SiteIdentifierProviderFunc`, `RestrictedSourcesProviderFunc`, `UserProviderFunc`, `IsAutomaticRedirectInitiallyStaleProviderFunc`.

## Core Business Logic

- **Cycle detection**: Validates redirects don't form loops; cyclic redirects are marked stale
- **Flattening**: Optimizes multi-hop redirect chains into single-target redirects
- **Consolidation**: Updates redirect targets when intermediate targets change
- **Auto-creation**: Generates redirects from content tree changes (contentserver integration)

## Conventions

- Commit messages follow Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, etc.)
- Releases triggered by `v*.*.*` tags via goreleaser (library-only, no binaries)
