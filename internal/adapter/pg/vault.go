package pg

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// GetVaultByVerification 读取验证请求对应的 active 保险库条目。
func (s *IdentityStore) GetVaultByVerification(ctx context.Context, developerID, verificationID uuid.UUID) (*gormmodel.IdentityVault, error) {
	var entry gormmodel.IdentityVault
	if err := s.db.WithContext(ctx).
		Where("developer_id = ? AND verification_id = ? AND status = 'active'", developerID, verificationID).
		Take(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// CreateVaultEntry 插入保险库条目。
func (s *IdentityStore) CreateVaultEntry(ctx context.Context, entry *gormmodel.IdentityVault) error {
	return s.db.WithContext(ctx).Create(entry).Error
}

// GetVaultByToken 按 vault token 与开发者读取条目。
func (s *IdentityStore) GetVaultByToken(ctx context.Context, developerID uuid.UUID, token string) (*gormmodel.IdentityVault, error) {
	var entry gormmodel.IdentityVault
	if err := s.db.WithContext(ctx).Where("developer_id = ? AND vault_token = ?", developerID, token).Take(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// ListVault 分页列出开发者保险库条目。
func (s *IdentityStore) ListVault(ctx context.Context, developerID uuid.UUID, page, limit int) ([]gormmodel.IdentityVault, int64, error) {
	var entries []gormmodel.IdentityVault
	var total int64
	query := s.db.WithContext(ctx).Model(&gormmodel.IdentityVault{}).Where("developer_id = ?", developerID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&entries).Error
	return entries, total, err
}

// DeleteVault 硬删除保险库条目。
func (s *IdentityStore) DeleteVault(ctx context.Context, developerID uuid.UUID, token string) (int64, error) {
	res := s.db.WithContext(ctx).Where("developer_id = ? AND vault_token = ?", developerID, token).Delete(&gormmodel.IdentityVault{})
	return res.RowsAffected, res.Error
}

// CreateShareLink 创建保险库分享链接。
func (s *IdentityStore) CreateShareLink(ctx context.Context, link *gormmodel.VaultShareLink) error {
	return s.db.WithContext(ctx).Create(link).Error
}

// GetShareLink 读取分享链接。
func (s *IdentityStore) GetShareLink(ctx context.Context, token string) (*gormmodel.VaultShareLink, error) {
	var link gormmodel.VaultShareLink
	if err := s.db.WithContext(ctx).Where("share_token = ?", token).Take(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

// IncrementShareAccess 原子增加分享访问次数，超过上限时返回 0。
func (s *IdentityStore) IncrementShareAccess(ctx context.Context, id uuid.UUID) (int64, error) {
	res := s.db.WithContext(ctx).Model(&gormmodel.VaultShareLink{}).
		Where("id = ? AND (max_accesses IS NULL OR access_count < max_accesses)", id).
		UpdateColumn("access_count", gorm.Expr("access_count + 1"))
	return res.RowsAffected, res.Error
}

// GetVaultByID 按主键读取保险库条目。
func (s *IdentityStore) GetVaultByID(ctx context.Context, id uuid.UUID) (*gormmodel.IdentityVault, error) {
	var entry gormmodel.IdentityVault
	if err := s.db.WithContext(ctx).Where("id = ?", id).Take(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// TouchVaultAccess 更新保险库最后访问时间。
func (s *IdentityStore) TouchVaultAccess(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&gormmodel.IdentityVault{}).Where("id = ?", id).
		Updates(map[string]any{"last_accessed_at": time.Now().UTC(), "access_count": gorm.Expr("access_count + 1")}).Error
}
