-- +goose Up
CREATE TABLE discovery_jobs (
    id                  TEXT PRIMARY KEY,
    tenant_id           TEXT NOT NULL,
    collector           TEXT NOT NULL,
    scope               JSONB NOT NULL DEFAULT '{}',
    mode                TEXT NOT NULL DEFAULT 'read_only',
    write_to_catalog    BOOLEAN NOT NULL DEFAULT true,
    write_to_memory     BOOLEAN NOT NULL DEFAULT false,
    requested_by        TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ
);

CREATE INDEX idx_discovery_jobs_tenant ON discovery_jobs (tenant_id, created_at DESC);

CREATE TABLE discovery_observations (
    id              TEXT PRIMARY KEY,
    job_id          TEXT NOT NULL REFERENCES discovery_jobs(id) ON DELETE CASCADE,
    collector       TEXT NOT NULL,
    observed_at     TIMESTAMPTZ NOT NULL,
    resource_ref    TEXT NOT NULL,
    kind            TEXT NOT NULL,
    claim           JSONB NOT NULL DEFAULT '{}',
    evidence        JSONB NOT NULL DEFAULT '{}',
    classification  TEXT,
    confidence      NUMERIC(5,4)
);

CREATE INDEX idx_discovery_observations_job ON discovery_observations (job_id);

-- +goose Down
DROP TABLE IF EXISTS discovery_observations;
DROP TABLE IF EXISTS discovery_jobs;
