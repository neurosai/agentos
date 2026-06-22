-- +goose Up
CREATE TABLE tasks (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL,
    context_id      TEXT,
    agent_ref       TEXT NOT NULL,
    status          TEXT NOT NULL,
    input           JSONB NOT NULL DEFAULT '{}',
    labels          JSONB NOT NULL DEFAULT '{}',
    created_by      TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_tenant_status ON tasks (tenant_id, status);
CREATE INDEX idx_tasks_created_at ON tasks (created_at DESC);

CREATE TABLE task_messages (
    id          TEXT PRIMARY KEY,
    task_id     TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    role        TEXT NOT NULL,
    content     TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_task_messages_task ON task_messages (task_id, created_at);

CREATE TABLE task_artifacts (
    id              TEXT PRIMARY KEY,
    task_id         TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    content_type    TEXT NOT NULL,
    uri             TEXT NOT NULL,
    size_bytes      BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_task_artifacts_task ON task_artifacts (task_id);

CREATE TABLE task_events (
    id          TEXT PRIMARY KEY,
    task_id     TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    event_type  TEXT NOT NULL,
    payload     JSONB NOT NULL DEFAULT '{}',
    trace_id    TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_task_events_task ON task_events (task_id, created_at);

CREATE TABLE task_approvals (
    id          TEXT PRIMARY KEY,
    task_id     TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    requested   TIMESTAMPTZ NOT NULL DEFAULT now(),
    decided_at  TIMESTAMPTZ,
    decided_by  TEXT,
    approved    BOOLEAN,
    reason      TEXT
);

CREATE INDEX idx_task_approvals_task ON task_approvals (task_id);

CREATE TABLE agent_runs (
    id          TEXT PRIMARY KEY,
    task_id     TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    agent_ref   TEXT NOT NULL,
    status      TEXT NOT NULL,
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at    TIMESTAMPTZ
);

CREATE INDEX idx_agent_runs_task ON agent_runs (task_id);

-- +goose Down
DROP TABLE IF EXISTS agent_runs;
DROP TABLE IF EXISTS task_approvals;
DROP TABLE IF EXISTS task_events;
DROP TABLE IF EXISTS task_artifacts;
DROP TABLE IF EXISTS task_messages;
DROP TABLE IF EXISTS tasks;
