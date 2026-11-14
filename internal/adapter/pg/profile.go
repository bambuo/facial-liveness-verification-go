package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"facial-liveness-verification-go/internal/domain"
)

const profileColumns = "id::text, tenant_id::text, subject_id::text, embedding_cipher, key_version, model_version, status, created_at, updated_at"

// ProfileRepo 是档案仓储适配器。
type ProfileRepo struct{ r *Repo }

// GetActive 查询主体的 active 档案。
func (pr *ProfileRepo) GetActive(ctx context.Context, tenantID, subjectID uuid.UUID) (*domain.Profile, error) {
	row := pr.r.Pool.QueryRow(ctx,
		"SELECT "+profileColumns+" FROM face_profiles WHERE tenant_id=$1::uuid AND subject_id=$2::uuid AND status='active'",
		tenantID.String(), subjectID.String())
	return scanProfile(row)
}

// Upsert 写入 active 档案；同主体已有 active 时覆盖特征与版本。
func (pr *ProfileRepo) Upsert(ctx context.Context, p *domain.Profile) (*domain.Profile, error) {
	_, err := pr.r.Pool.Exec(ctx, `
		INSERT INTO face_profiles (id, tenant_id, subject_id, embedding_cipher, key_version, model_version, status, created_at, updated_at)
		VALUES ($1::uuid,$2::uuid,$3::uuid,$4,$5,$6,'active',$7,$7)
		ON CONFLICT (subject_id) WHERE status='active'
		DO UPDATE SET embedding_cipher=EXCLUDED.embedding_cipher, key_version=EXCLUDED.key_version,
		              model_version=EXCLUDED.model_version, status='active', updated_at=EXCLUDED.updated_at`,
		p.ID.String(), p.TenantID.String(), p.SubjectID.String(), p.EmbeddingCipher, p.KeyVersion, p.ModelVersion, p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Update 更新档案状态。
func (pr *ProfileRepo) Update(ctx context.Context, p *domain.Profile) error {
	_, err := pr.r.Pool.Exec(ctx,
		"UPDATE face_profiles SET status=$2, updated_at=$3 WHERE tenant_id=$1::uuid AND id=$4::uuid",
		p.TenantID.String(), string(p.Status), p.UpdatedAt, p.ID.String())
	return err
}

// ListDeletedExpired 列出超过保留期的已删除档案。
func (pr *ProfileRepo) ListDeletedExpired(ctx context.Context, before time.Time, limit int) ([]*domain.Profile, error) {
	rows, err := pr.r.Pool.Query(ctx,
		"SELECT "+profileColumns+" FROM face_profiles WHERE status='deleted' AND updated_at < $1 ORDER BY updated_at LIMIT $2",
		before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Profile
	for rows.Next() {
		p, err := scanProfile(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Delete 硬删除档案。
func (pr *ProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := pr.r.Pool.Exec(ctx, "DELETE FROM face_profiles WHERE id=$1::uuid", id.String())
	return err
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProfile(row scanner) (*domain.Profile, error) {
	var (
		id, tenant, subject, modelVersion, status string
		cipher                                    []byte
		keyVersion                                int
		created, updated                          time.Time
	)
	err := row.Scan(&id, &tenant, &subject, &cipher, &keyVersion, &modelVersion, &status, &created, &updated)
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
	sid, err := uuid.Parse(subject)
	if err != nil {
		return nil, err
	}
	return &domain.Profile{ID: uid, TenantID: tid, SubjectID: sid, EmbeddingCipher: cipher, KeyVersion: keyVersion, ModelVersion: modelVersion, Status: domain.ProfileStatus(status), CreatedAt: created, UpdatedAt: updated}, nil
}
