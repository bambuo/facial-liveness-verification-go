package pg

import (
	"context"
	"time"

	"github.com/google/uuid"

	"gorm.io/gorm"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// MaintenanceStore 执行 idswyft 周期清理与一致性检查。
type MaintenanceStore struct {
	db *gorm.DB
}

// NewMaintenanceStore 构造维护仓储。
func NewMaintenanceStore(db *gorm.DB) *MaintenanceStore {
	return &MaintenanceStore{db: db}
}

// DemoCleanup 删除超过保留期的 sandbox 验证请求；关联子表由外键级联清理。
func (s *MaintenanceStore) DemoCleanup(ctx context.Context, before time.Time) (int64, error) {
	res := s.db.WithContext(ctx).Where("is_sandbox = TRUE AND created_at < ?", before).
		Delete(&gormmodel.VerificationRequest{})
	return res.RowsAffected, res.Error
}

// AbandonedCaptureCleanup 删除过期验证请求的文档与自拍记录。
func (s *MaintenanceStore) AbandonedCaptureCleanup(ctx context.Context, before time.Time) (int64, error) {
	var ids []uuid.UUID
	if err := s.db.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).
		Where("status IN ? AND created_at < ?", []string{"pending", "processing"}, before).
		Pluck("id", &ids).Error; err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	var total int64
	res := s.db.WithContext(ctx).Where("verification_request_id IN ?", ids).Delete(&gormmodel.Selfie{})
	total += res.RowsAffected
	if res.Error != nil {
		return total, res.Error
	}
	res = s.db.WithContext(ctx).Where("verification_request_id IN ?", ids).Delete(&gormmodel.Document{})
	total += res.RowsAffected
	return total, res.Error
}

// ActivityLogCleanup 删除超过保留期的 API 活动日志。
func (s *MaintenanceStore) ActivityLogCleanup(ctx context.Context, before time.Time) (int64, error) {
	res := s.db.WithContext(ctx).Where("created_at < ?", before).Delete(&gormmodel.APIActivityLog{})
	return res.RowsAffected, res.Error
}

// IdempotencyCleanup 删除已过期的幂等键。
func (s *MaintenanceStore) IdempotencyCleanup(ctx context.Context, now time.Time) (int64, error) {
	res := s.db.WithContext(ctx).Where("expires_at < ?", now).Delete(&gormmodel.IdempotencyKey{})
	return res.RowsAffected, res.Error
}

// WebhookPayloadCleanup 清空超过保留期的投递载荷，保留投递审计记录。
func (s *MaintenanceStore) WebhookPayloadCleanup(ctx context.Context, before time.Time) (int64, error) {
	res := s.db.WithContext(ctx).Model(&gormmodel.WebhookDelivery{}).
		Where("created_at < ? AND payload <> ?", before, []byte("{}")).
		Update("payload", []byte("{}"))
	return res.RowsAffected, res.Error
}

// ConsistencyCheck 返回活跃验证请求与待投递 Webhook 的数量，作为轻量一致性信号。
func (s *MaintenanceStore) ConsistencyCheck(ctx context.Context) (active, deliveries int64, err error) {
	if err = s.db.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).
		Where("status IN ?", []string{"pending", "processing"}).Count(&active).Error; err != nil {
		return 0, 0, err
	}
	err = s.db.WithContext(ctx).Model(&gormmodel.WebhookDelivery{}).
		Where("status = ?", "pending").Count(&deliveries).Error
	return active, deliveries, err
}
