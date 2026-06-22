# AgentOS Contracts

Version: **v0.1.0** (foundation)

## Northbound APIs

External APIs use **HTTP/JSON** (OpenAPI 3.1). Internal streaming uses **protobuf** over Connect/gRPC.

| Daemon | Spec file | Base URL (Compose) |
|--------|-----------|-------------------|
| TaskD | `api/openapi/taskd.yaml` | `:8081` |
| PolicyD | `api/openapi/policyd.yaml` | `:8082` |
| AuditD | `api/openapi/auditd.yaml` | `:8083` |
| ToolD | `api/openapi/toold.yaml` | `:8084` |
| MemoryD | `api/openapi/memoryd.yaml` | `:8085` |
| CatalogD | `api/openapi/catalogd.yaml` | `:8086` |
| DiscoveryD | `api/openapi/discoveryd.yaml` | `:8087` |

Shared components: `api/openapi/common.yaml`

**Deferred:** `agentd.yaml` (Agent manifest and runtime APIs)

## Internal protobuf

| Service | Proto file |
|---------|------------|
| `TaskEventsService.StreamTaskEvents` | `proto/agentos/v1/task_events.proto` |
| `CatalogIngestService.UpsertObservations` | `proto/agentos/v1/catalog_ingest.proto` |

Generate with: `buf generate`

## JSON Schemas

| Schema | Purpose |
|--------|---------|
| `memory-record.json` | MemoryD write requests |
| `catalog-entity.json` | CatalogD entity documents |
| `agent-manifest.json` | AgentD manifests (`x-deferred: agentd`) |

Examples in `examples/` are validated in CI via `go test -tags contracts`.

## Error model

All APIs return a canonical error envelope:

```json
{
  "error": {
    "code": "policy_denied",
    "message": "tool invocation denied by policy",
    "requestId": "req_01J...",
    "retryable": false,
    "details": {}
  }
}
```

Codes: `invalid_request`, `unauthenticated`, `forbidden`, `not_found`, `conflict`, `policy_denied`, `rate_limited`, `backend_unavailable`

## Compatibility matrix (stub)

| AgentOS | MCP | A2A | Storage |
|---------|-----|-----|---------|
| v0.1 | TBD | TBD | PostgreSQL + pgvector |

## Versioning rules

- Spec version (`v0.1`, `v0.2`) is independent of runtime release tags
- Breaking API changes require a new OpenAPI major path (`/v2/`)
- Protobuf packages use `agentos.v1`; breaking changes increment package version
- Database migrations are forward-only via goose; rollbacks defined per file
