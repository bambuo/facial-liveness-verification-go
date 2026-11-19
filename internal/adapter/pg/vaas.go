package pg

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// CreateVAASUser 插入用户。
func (s *IdentityStore) CreateVAASUser(ctx context.Context, user *gormmodel.User) error {
	return s.db.WithContext(ctx).Create(user).Error
}

// EnsureUser 仅在用户不存在时创建用户行。
func (s *IdentityStore) EnsureUser(ctx context.Context, id uuid.UUID) error {
	user := &gormmodel.User{ID: id}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(user).Error
}

// ListDevelopers 搜索并分页开发者。
func (s *IdentityStore) ListDevelopers(ctx context.Context, search, status string, page, limit int) ([]gormmodel.Developer, int64, error) {
	var developers []gormmodel.Developer
	var total int64
	query := s.db.WithContext(ctx).Model(&gormmodel.Developer{})
	if strings.TrimSpace(search) != "" {
		term := "%" + strings.TrimSpace(search) + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR company ILIKE ?", term, term, term)
	}
	if status == "active" || status == "suspended" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&developers).Error
	return developers, total, err
}

// SetDeveloperStatus 修改开发者状态。
func (s *IdentityStore) SetDeveloperStatus(ctx context.Context, id uuid.UUID, status string) (int64, error) {
	res := s.db.WithContext(ctx).Model(&gormmodel.Developer{}).Where("id = ?", id).Update("status", status)
	return res.RowsAffected, res.Error
}

// CountAPIKeys 统计开发者 API Key 数量。
func (s *IdentityStore) CountAPIKeys(ctx context.Context, developerID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&gormmodel.APIKey{}).Where("developer_id = ?", developerID).Count(&count).Error
	return count, err
}

// CountVerifications 统计开发者验证数量。
func (s *IdentityStore) CountVerifications(ctx context.Context, developerID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).Where("developer_id = ?", developerID).Count(&count).Error
	return count, err
}
