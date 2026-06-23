package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/neurosai/agentos/internal/domain/memory"
	"github.com/neurosai/agentos/internal/port"
)

// MemoryRepository stores memory records in PostgreSQL.
type MemoryRepository struct {
	pool *Pool
}

func NewMemoryRepository(pool *Pool) *MemoryRepository {
	return &MemoryRepository{pool: pool}
}

func (r *MemoryRepository) Create(ctx context.Context, rec memory.Record) error {
	contentJSON, _ := json.Marshal(rec.ContentJSON)
	provenance, _ := json.Marshal(rec.Provenance)
	acl, _ := json.Marshal(rec.ACL)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO memory_records (
			id, tenant_id, namespace, type, subject_ref, content, content_json,
			classification, confidence, source_type, source_ref, provenance, acl,
			retention_until, created_by, created_at, supersedes, deleted_at
		) VALUES ($1::uuid,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,NULL,NULL)`,
		uuidFromID(rec.ID), rec.TenantID, rec.Namespace, string(rec.Type), rec.SubjectRef,
		rec.Content, contentJSON, rec.Classification, rec.Confidence, rec.SourceType, rec.SourceRef,
		provenance, acl, rec.RetentionUntil, rec.CreatedBy, rec.CreatedAt)
	return err
}

func (r *MemoryRepository) Get(ctx context.Context, id string) (memory.Record, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id::text, tenant_id, namespace, type, subject_ref, content, content_json,
			classification, confidence, source_type, source_ref, provenance, acl,
			retention_until, created_by, created_at, supersedes::text, deleted_at
		FROM memory_records WHERE id = $1::uuid`, uuidFromID(id))
	return scanMemory(row)
}

func (r *MemoryRepository) SoftDelete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE memory_records SET deleted_at = $1 WHERE id = $2::uuid`,
		time.Now().UTC(), uuidFromID(id))
	return err
}

func (r *MemoryRepository) Revise(ctx context.Context, id string, revised memory.Record) error {
	return r.Create(ctx, revised)
}

func (r *MemoryRepository) Query(ctx context.Context, q memory.Query, _ []memory.Record) ([]memory.QueryResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	query := `
		SELECT id::text, tenant_id, namespace, type, subject_ref, content, content_json,
			classification, confidence, source_type, source_ref, provenance, acl,
			retention_until, created_by, created_at, supersedes::text, deleted_at
		FROM memory_records
		WHERE deleted_at IS NULL`
	args := []any{}
	n := 1
	if q.Namespace != "" {
		query += fmt.Sprintf(" AND namespace = $%d", n)
		args = append(args, q.Namespace)
		n++
	}
	if q.QueryText != "" {
		query += fmt.Sprintf(" AND content ILIKE $%d", n)
		args = append(args, "%"+q.QueryText+"%")
		n++
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", n)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []memory.QueryResult
	for rows.Next() {
		rec, err := scanMemoryRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, memory.QueryResult{Record: rec, Score: 1.0})
	}
	return out, rows.Err()
}

var _ port.MemoryRepository = (*MemoryRepository)(nil)

func scanMemory(row pgx.Row) (memory.Record, error) {
	return scanMemoryRow(row)
}

type scannable interface {
	Scan(dest ...any) error
}

func scanMemoryRow(row scannable) (memory.Record, error) {
	var rec memory.Record
	var typ string
	var contentJSON, provenance, acl []byte
	var supersedes *string
	err := row.Scan(
		&rec.ID, &rec.TenantID, &rec.Namespace, &typ, &rec.SubjectRef, &rec.Content, &contentJSON,
		&rec.Classification, &rec.Confidence, &rec.SourceType, &rec.SourceRef, &provenance, &acl,
		&rec.RetentionUntil, &rec.CreatedBy, &rec.CreatedAt, &supersedes, &rec.DeletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return memory.Record{}, fmt.Errorf("memory record not found")
	}
	if err != nil {
		return memory.Record{}, err
	}
	rec.Type = memory.Type(typ)
	_ = json.Unmarshal(contentJSON, &rec.ContentJSON)
	_ = json.Unmarshal(provenance, &rec.Provenance)
	_ = json.Unmarshal(acl, &rec.ACL)
	if supersedes != nil {
		rec.Supersedes = *supersedes
	}
	return rec, nil
}

func uuidFromID(id string) string {
	id = strings.TrimPrefix(id, "mem_")
	if len(id) == 36 {
		return id
	}
	// store deterministic uuid in text id suffix for v0.2 — use raw id if valid uuid
	return id
}
