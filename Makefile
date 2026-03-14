.DEFAULT_GOAL:=help
-include .makerc

# --- Config -----------------------------------------------------------------

# Newline hack for error output
define br


endef

# --- Targets -----------------------------------------------------------------

# This allows us to accept extra arguments
%: .mise .lefthook
	@:

.PHONY: .lefthook
# Configure git hooks for lefthook
.lefthook:
	@lefthook install --reset-hooks-path

.PHONY: .mise
# Install dependencies
.mise:
ifeq (, $(shell command -v mise))
	$(error $(br)$(br)Please ensure you have 'mise' installed and activated!$(br)$(br)  $$ brew update$(br)  $$ brew install mise$(br)$(br)See the documentation: https://mise.jdx.dev/getting-started.html)
endif
	@mise install

### Tasks

.PHONY: tidy
## Run go mod tidy
tidy:
	@echo "〉running go mod tidy"
	@go mod tidy

.PHONY: generate
## Run go generate
generate:
	@echo "〉running go generate"
	@go generate ./...

.PHONY: test
## Run tests
test:
	@echo "〉running tests"
	@go test -coverprofile=coverage.out ./...

.PHONY: test.race
## Run tests with `-race` flag
test.race:
	@echo "〉running tests with -race flag"
	@go test -race -coverprofile=coverage.out ./...

.PHONY: test.cover
## Run tests with coverage
test.cover:
	@echo "〉running tests with coverage"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out

.PHONY: fmt
## Run format
fmt:
	@echo "〉formatting files"
	@golangci-lint fmt ./...

.PHONY: lint
## Run linter
lint:
	@echo "〉linting files"
	@golangci-lint run

.PHONY: lint.fix
## Fix lint violations
lint.fix:
	@echo "〉fixing lint errors"
	@golangci-lint run --fix

.PHONY: outdated
## Show outdated direct dependencies
outdated:
	@echo "〉listing outdated dependencies"
	@go list -u -m -json all | go-mod-outdated -update -direct

### Documentation

.PHONY: doc
## Open go docs
doc:
	@echo "〉starting go docs"
	@open "http://localhost:6060/pkg/github.com/foomo/redirects/v2/"
	@godoc -http=localhost:6060 -play

### Utils

.PHONY: help
## Show help text
help:
	@echo ""
	@echo "Welcome to redirects!"
	@echo "\nUsage:\n  make [task]"
	@awk '{ \
		if($$0 ~ /^### /){ \
			if(help) printf "%-23s %s\n\n", cmd, help; help=""; \
			printf "\n%s:\n", substr($$0,5); \
		} else if($$0 ~ /^[a-zA-Z0-9._-]+:/){ \
			cmd = substr($$0, 1, index($$0, ":")-1); \
			if(help) printf "  %-23s %s\n", cmd, help; help=""; \
		} else if($$0 ~ /^##/){ \
			help = help ? help "\n                          " substr($$0,3) : substr($$0,3); \
		} else if(help){ \
			print "\n                        " help "\n"; help=""; \
		} \
	}' $(MAKEFILE_LIST)
