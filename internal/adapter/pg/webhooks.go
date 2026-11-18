package pg

import (
	"context"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// WebhooksByDeveloper 列出开发者的 Webhook。
func (s *IdentityStore) WebhooksByDeveloper(ctx context.Context, developerID uuid.UUID) ([]gormmodel.Webhook, error) {
	var webhooks []gormmodel.Webhook
	err := s.db.WithContext(ctx).Where("developer_id = ?", developerID).Order("created_at DESC").Find(&webhooks).Error
	return webhooks, err
}

// WebhookByID 读取开发者所属 Webhook。
func (s *IdentityStore) WebhookByID(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.Webhook, error) {
	var webhook gormmodel.Webhook
	if err := s.db.WithContext(ctx).Where("id = ? AND developer_id = ?", id, developerID).Take(&webhook).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

// WebhookByURL 检查同开发者同 URL 的 Webhook。
func (s *IdentityStore) WebhookByURL(ctx context.Context, developerID uuid.UUID, target string, sandbox bool) (*gormmodel.Webhook, error) {
	var webhook gormmodel.Webhook
	if err := s.db.WithContext(ctx).Where("developer_id = ? AND url = ? AND is_sandbox = ?", developerID, target, sandbox).Take(&webhook).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

// CreateWebhook 创建 Webhook。
func (s *IdentityStore) CreateWebhook(ctx context.Context, webhook *gormmodel.Webhook) error {
	return s.db.WithContext(ctx).Create(webhook).Error
}

// UpdateWebhook 更新开发者 Webhook。
func (s *IdentityStore) UpdateWebhook(ctx context.Context, developerID, id uuid.UUID, updates map[string]any) (int64, error) {
	res := s.db.WithContext(ctx).Model(&gormmodel.Webhook{}).Where("id = ? AND developer_id = ?", id, developerID).Updates(updates)
	return res.RowsAffected, res.Error
}

// DeleteWebhook 删除开发者 Webhook。
func (s *IdentityStore) DeleteWebhook(ctx context.Context, developerID, id uuid.UUID) (int64, error) {
	res := s.db.WithContext(ctx).Where("id = ? AND developer_id = ?", id, developerID).Delete(&gormmodel.Webhook{})
	return res.RowsAffected, res.Error
}

// WebhookDeliveries 分页列出投递记录。
func (s *IdentityStore) WebhookDeliveries(ctx context.Context, webhookID uuid.UUID, page, limit int) ([]gormmodel.WebhookDelivery, int64, error) {
	var deliveries []gormmodel.WebhookDelivery
	var total int64
	query := s.db.WithContext(ctx).Model(&gormmodel.WebhookDelivery{}).Where("webhook_id = ?", webhookID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&deliveries).Error
	return deliveries, total, err
}
