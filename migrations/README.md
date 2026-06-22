# Database Migrations

AgentOS uses [goose](https://github.com/pressly/goose) for PostgreSQL migrations.

## Extensions

- `pgcrypto` — UUID generation
- `vector` — pgvector for memory embeddings (default dimension: 1536)

## Entity relationship

```mermaid
erDiagram
    tasks ||--o{ task_messages : has
    tasks ||--o{ task_artifacts : has
    tasks ||--o{ task_events : has
    tasks ||--o{ task_approvals : has
    tasks ||--o{ agent_runs : has
    discovery_jobs ||--o{ discovery_observations : produces
    catalog_entities ||--o{ catalog_edges : from
    catalog_entities ||--o{ catalog_edges : to
    tool_registry ||--o{ tool_invocations : tracks
```

## Retention defaults

| Data | Default retention |
|------|------------------:|
| Task events | 30 days |
| Audit events | 180 days |
| Discovery observations | 90 days |
| Evidence memory | 90 days |
| Task memory | 30 days |
| Session memory | 7 days |

## Commands

```bash
export DATABASE_URL="postgres://agentos:agentos@localhost:5432/agentos?sslmode=disable"
goose -dir migrations postgres "$DATABASE_URL" up
goose -dir migrations postgres "$DATABASE_URL" down
```

## Embedding dimension

Migration `00004_memory.sql` uses `vector(1536)`. To change dimensions, add a new migration that alters the column and rebuilds the HNSW index.
