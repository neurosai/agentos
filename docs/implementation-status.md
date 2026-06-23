# Implementation Status

Factual matrix of AgentOS v0.2 components. Update when behavior changes; do not use as a devlog.

| Component | API contract | Domain model | Module | Postgres adapter | Status |
|-----------|--------------|--------------|--------|------------------|--------|
| agentosd | — | — | `cmd/agentosd` | — | implemented |
| agentctl | — | — | `cmd/agentctl` | — | implemented |
| TaskD | `api/openapi/taskd.yaml` | `internal/domain/task` | `internal/module/task` | `internal/adapter/postgres/taskrepo.go` | implemented |
| AuditD | `api/openapi/auditd.yaml` | `internal/domain/audit` | `internal/module/audit` | `internal/adapter/postgres/auditrepo.go` | implemented |
| PolicyD | `api/openapi/policyd.yaml` | `internal/domain/policy` | `internal/module/policy` | OPA via `internal/adapter/opa` | implemented |
| ToolD | `api/openapi/toold.yaml` | `internal/domain/tool` | `internal/module/tool` | `internal/adapter/postgres/toolrepo.go` | implemented |
| MemoryD | `api/openapi/memoryd.yaml` | `internal/domain/memory` | `internal/module/memory` | `internal/adapter/postgres/memoryrepo.go` | implemented |
| CatalogD | `api/openapi/catalogd.yaml` | `internal/domain/catalog` | — | — | planned |
| DiscoveryD | `api/openapi/discoveryd.yaml` | `internal/domain/discovery` | — | — | planned |
| AgentD | — | — | — | — | planned (v0.3) |

## Legacy daemon stubs

`cmd/taskd`, `cmd/policyd`, `cmd/auditd`, `cmd/toold`, `cmd/memoryd`, `cmd/catalogd`, `cmd/discoveryd` remain build-only placeholders until split-mode deployment (v0.4+). Use `agentosd serve` for v0.2.

## Deferred

- OIDC authentication (dev stub bearer token in v0.2)
- Embedding generation and Qdrant
- Hermes / AgentD runtime
- Nix packaging
- Vault credential exchange
