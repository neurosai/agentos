-- +goose Up
ALTER TABLE tool_invocations ADD COLUMN IF NOT EXISTS idempotency_key TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tool_invocations_idempotency
    ON tool_invocations (tool_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key != '';

-- +goose Down
DROP INDEX IF EXISTS idx_tool_invocations_idempotency;
ALTER TABLE tool_invocations DROP COLUMN IF EXISTS idempotency_key;
