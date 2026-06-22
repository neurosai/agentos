-- +goose Up
CREATE TABLE catalog_entities (
    ref         TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL,
    kind        TEXT NOT NULL,
    namespace   TEXT,
    name        TEXT NOT NULL,
    title       TEXT,
    labels      JSONB NOT NULL DEFAULT '{}',
    spec        JSONB NOT NULL DEFAULT '{}',
    source      TEXT NOT NULL DEFAULT 'declared',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_catalog_entities_tenant_kind ON catalog_entities (tenant_id, kind);
CREATE INDEX idx_catalog_entities_name ON catalog_entities (name);

CREATE TABLE catalog_edges (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       TEXT NOT NULL,
    from_ref        TEXT NOT NULL REFERENCES catalog_entities(ref) ON DELETE CASCADE,
    to_ref          TEXT NOT NULL REFERENCES catalog_entities(ref) ON DELETE CASCADE,
    relation_type   TEXT NOT NULL,
    confidence      NUMERIC(5,4),
    observed_at     TIMESTAMPTZ,
    declared        BOOLEAN NOT NULL DEFAULT false,
    UNIQUE (from_ref, to_ref, relation_type)
);

CREATE INDEX idx_catalog_edges_from ON catalog_edges (from_ref);
CREATE INDEX idx_catalog_edges_to ON catalog_edges (to_ref);

CREATE TABLE catalog_snapshots (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL,
    root_ref    TEXT NOT NULL,
    entity_ids  JSONB NOT NULL DEFAULT '[]',
    edge_count  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS catalog_snapshots;
DROP TABLE IF EXISTS catalog_edges;
DROP TABLE IF EXISTS catalog_entities;
