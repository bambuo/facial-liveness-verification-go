package pg

import (
	"context"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// DeveloperAPIKeys 列出开发者的 API Key。
func (s *IdentityStore) DeveloperAPIKeys(ctx context.Context, developerID uuid.UUID) ([]gormmodel.APIKey, error) {
	var keys []gormmodel.APIKey
	err := s.db.WithContext(ctx).Where("developer_id = ?", developerID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

// CreateDeveloperAPIKey 创建 API Key。
func (s *IdentityStore) CreateDeveloperAPIKey(ctx context.Context, key *gormmodel.APIKey) error {
	return s.db.WithContext(ctx).Create(key).Error
}

// RevokeDeveloperAPIKey 撤销开发者 API Key。
func (s *IdentityStore) RevokeDeveloperAPIKey(ctx context.Context, developerID, id uuid.UUID) (int64, error) {
	res := s.db.WithContext(ctx).Model(&gormmodel.APIKey{}).
		Where("id = ? AND developer_id = ? AND revoked_at IS NULL", id, developerID).
		Updates(map[string]any{"is_active": false, "revoked_at": time.Now().UTC(), "revoked_reason": "developer request"})
	return res.RowsAffected, res.Error
}

// GetDeveloperAPIKey 读取开发者 API Key。
func (s *IdentityStore) GetDeveloperAPIKey(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.APIKey, error) {
	var key gormmodel.APIKey
	if err := s.db.WithContext(ctx).Where("id = ? AND developer_id = ?", id, developerID).Take(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}
