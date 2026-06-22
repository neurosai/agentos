-- +goose Up
CREATE TABLE tool_registry (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    name            TEXT NOT NULL,
    transport       TEXT NOT NULL,
    mcp_server      TEXT,
    risk            TEXT NOT NULL DEFAULT 'low',
    description     TEXT,
    schema_json     JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tool_registry_tenant ON tool_registry (tenant_id);

CREATE TABLE tool_invocations (
    id              TEXT PRIMARY KEY,
    tool_id         TEXT NOT NULL REFERENCES tool_registry(id),
    task_id         TEXT,
    agent_id        TEXT,
    status          TEXT NOT NULL,
    audit_event_id  TEXT,
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX idx_tool_invocations_tool ON tool_invocations (tool_id, started_at DESC);
CREATE INDEX idx_tool_invocations_task ON tool_invocations (task_id);

-- +goose Down
DROP TABLE IF EXISTS tool_invocations;
DROP TABLE IF EXISTS tool_registry;
