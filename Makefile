.PHONY: all build test lint verify generate proto lint-proto policy migrate-up migrate-down contracts clean

GO ?= go
BUF ?= buf
GOOSE ?= goose
OPA ?= opa
GOLANGCI_LINT ?= golangci-lint

MODULE := github.com/neurosai/agentos
BINARIES := agentctl taskd policyd auditd toold memoryd catalogd discoveryd

all: build

build:
	@for bin in $(BINARIES); do \
		$(GO) build -o bin/$$bin ./cmd/$$bin; \
	done

test:
	$(GO) test ./...

lint:
	$(GOLANGCI_LINT) run ./...

lint-proto:
	$(BUF) lint

generate:
	$(GO) generate ./...

proto:
	$(BUF) generate

policy:
	$(OPA) test ./policies/

migrate-up:
	$(GOOSE) -dir migrations postgres "$${DATABASE_URL}" up

migrate-down:
	$(GOOSE) -dir migrations postgres "$${DATABASE_URL}" down

contracts:
	$(GO) test -tags contracts ./internal/contracts/...

verify: lint lint-proto generate test policy contracts
	@echo "verify: all checks passed"

clean:
	rm -rf bin/
	$(GO) clean -testcache
