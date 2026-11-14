package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"facial-liveness-verification-go/internal/domain"
)

// AuditRepo 是审计仓储适配器。
type AuditRepo struct{ r *Repo }

// DocumentRepo 是证件记录仓储适配器。
type DocumentRepo struct{ r *Repo }

// KeyRepo 是租户密钥密文仓储适配器。
type KeyRepo struct{ r *Repo }

// FaceIndex 是 pgvector 成员清理适配器。
type FaceIndex struct{ r *Repo }

// Add 写入审计并回填自增 ID。
func (ar *AuditRepo) Add(ctx context.Context, log *domain.AuditLog) error {
	return ar.r.Pool.QueryRow(ctx, `
		INSERT INTO face_audit_logs (tenant_id, actor, action, resource, result, request_id, detail_json, created_at)
		VALUES ($1::uuid,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		log.TenantID.String(), log.Actor, log.Action, log.Resource, log.Result, log.RequestID, log.DetailJSON, log.CreatedAt).Scan(&log.ID)
}

// ListByResource 按资源列出审计。
func (ar *AuditRepo) ListByResource(ctx context.Context, tenantID uuid.UUID, resource string) ([]*domain.AuditLog, error) {
	rows, err := ar.r.Pool.Query(ctx, `
		SELECT id, tenant_id::text, actor, action, resource, result, request_id, detail_json, created_at
		FROM face_audit_logs WHERE tenant_id=$1::uuid AND resource=$2 ORDER BY created_at`,
		tenantID.String(), resource)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.AuditLog
	for rows.Next() {
		var (
			id        int64
			tenant    string
			requestID *string
			detail    []byte
			created   time.Time
			l         domain.AuditLog
		)
		if err := rows.Scan(&id, &tenant, &l.Actor, &l.Action, &l.Resource, &l.Result, &requestID, &detail, &created); err != nil {
			return nil, err
		}
		tid, err := uuid.Parse(tenant)
		if err != nil {
			return nil, err
		}
		l.ID, l.TenantID, l.DetailJSON, l.CreatedAt, l.RequestID = id, tid, detail, created, requestID
		out = append(out, &l)
	}
	return out, rows.Err()
}

// Update 更新审计详情。
func (ar *AuditRepo) Update(ctx context.Context, log *domain.AuditLog) error {
	_, err := ar.r.Pool.Exec(ctx, "UPDATE face_audit_logs SET detail_json=$2 WHERE id=$1", log.ID, log.DetailJSON)
	return err
}

// ListExpired 列出留存到期的证件记录。
func (dr *DocumentRepo) ListExpired(ctx context.Context, now time.Time, limit int) ([]*domain.Document, error) {
	rows, err := dr.r.Pool.Query(ctx,
		"SELECT id::text, tenant_id::text, retention_until FROM face_documents WHERE retention_until IS NOT NULL AND retention_until < $1 ORDER BY retention_until LIMIT $2",
		now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Document
	for rows.Next() {
		var id, tenant string
		var until time.Time
		if err := rows.Scan(&id, &tenant, &until); err != nil {
			return nil, err
		}
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		tid, err := uuid.Parse(tenant)
		if err != nil {
			return nil, err
		}
		out = append(out, &domain.Document{ID: uid, TenantID: tid, RetentionUntil: &until})
	}
	return out, rows.Err()
}

// Delete 删除证件记录。
func (dr *DocumentRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := dr.r.Pool.Exec(ctx, "DELETE FROM face_documents WHERE id=$1::uuid", id.String())
	return err
}

// FindActive 返回租户最新 active 密钥密文。
func (kr *KeyRepo) FindActive(ctx context.Context, tenantID uuid.UUID) ([]byte, int, error) {
	var cipher []byte
	var version int
	err := kr.r.Pool.QueryRow(ctx,
		"SELECT key_cipher, key_version FROM face_tenant_keys WHERE tenant_id=$1::uuid AND status='active' ORDER BY key_version DESC LIMIT 1",
		tenantID.String()).Scan(&cipher, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	return cipher, version, nil
}

// MaxVersion 返回租户当前最大密钥版本。
func (kr *KeyRepo) MaxVersion(ctx context.Context, tenantID uuid.UUID) (int, error) {
	var version int
	err := kr.r.Pool.QueryRow(ctx, "SELECT COALESCE(MAX(key_version),0) FROM face_tenant_keys WHERE tenant_id=$1::uuid", tenantID.String()).Scan(&version)
	return version, err
}

// Insert 写入包裹后的租户密钥；版本冲突时忽略。
func (kr *KeyRepo) Insert(ctx context.Context, tenantID uuid.UUID, version int, cipher []byte) error {
	_, err := kr.r.Pool.Exec(ctx, `
		INSERT INTO face_tenant_keys (id, tenant_id, key_version, key_cipher, status, created_at)
		VALUES ($1::uuid,$2::uuid,$3,$4,'active',now()) ON CONFLICT (tenant_id, key_version) DO NOTHING`,
		uuid.New().String(), tenantID.String(), version, cipher)
	return err
}

// DeleteSubject 删除主体在租户全部 gallery 的向量成员。
func (fi *FaceIndex) DeleteSubject(ctx context.Context, tenantID, subjectID uuid.UUID) error {
	_, err := fi.r.Pool.Exec(ctx, `
		DELETE FROM face_gallery_members WHERE subject_id=$1::uuid
		AND gallery_id IN (SELECT id FROM face_galleries WHERE tenant_id=$2::uuid)`,
		subjectID.String(), tenantID.String())
	return err
}
