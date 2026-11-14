package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"facial-liveness-verification-go/internal/domain"
)

// TenantRepo 是租户仓储适配器。
type TenantRepo struct{ r *Repo }

// SubjectRepo 是主体仓储适配器。
type SubjectRepo struct{ r *Repo }

// GetByID 按租户 UUID 查询。
func (t *TenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return t.scanTenant(ctx, "SELECT id::text, code, name, status, policy_json, created_at, updated_at FROM face_tenants WHERE id=$1::uuid", id.String())
}

// GetByCode 按租户 code 查询。
func (t *TenantRepo) GetByCode(ctx context.Context, code string) (*domain.Tenant, error) {
	return t.scanTenant(ctx, "SELECT id::text, code, name, status, policy_json, created_at, updated_at FROM face_tenants WHERE code=$1", code)
}

func (t *TenantRepo) scanTenant(ctx context.Context, sql string, arg any) (*domain.Tenant, error) {
	var (
		id, code, name, status string
		policy                 []byte
		created, updated       time.Time
	)
	err := t.r.Pool.QueryRow(ctx, sql, arg).Scan(&id, &code, &name, &status, &policy, &created, &updated)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return &domain.Tenant{
		ID:         uid,
		Code:       code,
		Name:       name,
		Status:     domain.TenantStatus(status),
		PolicyJSON: policy,
		CreatedAt:  created,
		UpdatedAt:  updated,
	}, nil
}

const subjectColumns = "id::text, tenant_id::text, external_subject_id, status, created_at, updated_at, deleted_at"

// Get 按主体 UUID 查询。
func (s *SubjectRepo) Get(ctx context.Context, tenantID, subjectID uuid.UUID) (*domain.Subject, error) {
	return s.scanSubject(ctx, "SELECT "+subjectColumns+" FROM face_subjects WHERE tenant_id=$1::uuid AND id=$2::uuid", tenantID.String(), subjectID.String())
}

// GetByExternal 按业务用户 ID 查询；跨租户返回 nil。
func (s *SubjectRepo) GetByExternal(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.Subject, error) {
	return s.scanSubject(ctx, "SELECT "+subjectColumns+" FROM face_subjects WHERE tenant_id=$1::uuid AND external_subject_id=$2", tenantID.String(), externalID)
}

// GetOrCreate 取主体，不存在则创建；已删除主体恢复为 active。
func (s *SubjectRepo) GetOrCreate(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.Subject, error) {
	sub, err := s.GetByExternal(ctx, tenantID, externalID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if sub != nil {
		if sub.Status == domain.SubjectDeleted {
			_, err = s.r.Pool.Exec(ctx,
				"UPDATE face_subjects SET status='active', deleted_at=NULL, updated_at=$3 WHERE tenant_id=$1::uuid AND id=$2::uuid",
				tenantID.String(), sub.ID.String(), now)
			if err != nil {
				return nil, err
			}
			sub.Status = domain.SubjectActive
			sub.DeletedAt = nil
			sub.UpdatedAt = now
		}
		return sub, nil
	}
	id := uuid.New()
	_, err = s.r.Pool.Exec(ctx,
		"INSERT INTO face_subjects (id, tenant_id, external_subject_id, status, created_at, updated_at) VALUES ($1::uuid,$2::uuid,$3,'active',$4,$4) ON CONFLICT (tenant_id, external_subject_id) DO NOTHING",
		id.String(), tenantID.String(), externalID, now)
	if err != nil {
		return nil, err
	}
	return s.GetByExternal(ctx, tenantID, externalID)
}

// Update 更新主体状态。
func (s *SubjectRepo) Update(ctx context.Context, sub *domain.Subject) error {
	_, err := s.r.Pool.Exec(ctx,
		"UPDATE face_subjects SET status=$3, deleted_at=$4, updated_at=$5 WHERE tenant_id=$1::uuid AND id=$2::uuid",
		sub.TenantID.String(), sub.ID.String(), string(sub.Status), sub.DeletedAt, sub.UpdatedAt)
	return err
}

func (s *SubjectRepo) scanSubject(ctx context.Context, sql string, args ...any) (*domain.Subject, error) {
	var (
		id, tenant, external, status string
		created, updated             time.Time
		deleted                      *time.Time
	)
	err := s.r.Pool.QueryRow(ctx, sql, args...).Scan(&id, &tenant, &external, &status, &created, &updated, &deleted)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
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
	return &domain.Subject{ID: uid, TenantID: tid, ExternalID: external, Status: domain.SubjectStatus(status), CreatedAt: created, UpdatedAt: updated, DeletedAt: deleted}, nil
}
