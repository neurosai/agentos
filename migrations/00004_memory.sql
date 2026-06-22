-- +goose Up
-- Default embedding dimension: 1536 (OpenAI ada-002 class). Adjust via migration if needed.
CREATE TABLE memory_records (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id           TEXT NOT NULL,
    namespace           TEXT NOT NULL,
    type                TEXT NOT NULL,
    subject_ref         TEXT,
    content               TEXT NOT NULL,
    content_json          JSONB NOT NULL DEFAULT '{}',
    embedding             vector(1536),
    classification        TEXT,
    confidence            NUMERIC(5,4),
    source_type           TEXT,
    source_ref            TEXT NOT NULL,
    provenance            JSONB NOT NULL DEFAULT '{}',
    acl                   JSONB NOT NULL DEFAULT '{}',
    retention_until       TIMESTAMPTZ,
    created_by            TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    supersedes            UUID REFERENCES memory_records(id),
    deleted_at            TIMESTAMPTZ
);

CREATE INDEX idx_memory_tenant_namespace ON memory_records (tenant_id, namespace) WHERE deleted_at IS NULL;
CREATE INDEX idx_memory_type ON memory_records (type) WHERE deleted_at IS NULL;
CREATE INDEX idx_memory_retention ON memory_records (retention_until) WHERE deleted_at IS NULL;

-- HNSW index stub for approximate nearest neighbor search
CREATE INDEX idx_memory_embedding_hnsw ON memory_records
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);

-- +goose Down
DROP TABLE IF EXISTS memory_records;
