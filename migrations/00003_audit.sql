-- +goose Up
CREATE TABLE audit_events (
    event_id        TEXT PRIMARY KEY,
    occurred_at     TIMESTAMPTZ NOT NULL,
    tenant_id       TEXT NOT NULL,
    subject_id      TEXT,
    agent_id        TEXT,
    task_id         TEXT,
    event_type      TEXT NOT NULL,
    resource_type   TEXT,
    resource_id     TEXT,
    action          TEXT,
    decision        TEXT,
    status          TEXT,
    payload_hash    TEXT NOT NULL,
    prev_hash       TEXT,
    event_hash      TEXT NOT NULL,
    trace_id        TEXT,
    span_id         TEXT
);

CREATE INDEX idx_audit_tenant_occurred ON audit_events (tenant_id, occurred_at DESC);
CREATE INDEX idx_audit_task ON audit_events (task_id);
CREATE INDEX idx_audit_trace ON audit_events (trace_id);

-- +goose Down
DROP TABLE IF EXISTS audit_events;
