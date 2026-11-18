package pg

import (
	"context"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// CreateSchedule 创建重验证计划。
func (s *IdentityStore) CreateSchedule(ctx context.Context, schedule *gormmodel.ReverificationSchedule) error {
	return s.db.WithContext(ctx).Create(schedule).Error
}

// ListSchedules 分页列出开发者的重验证计划。
func (s *IdentityStore) ListSchedules(ctx context.Context, developerID uuid.UUID, status string, page, limit int) ([]gormmodel.ReverificationSchedule, int64, error) {
	var schedules []gormmodel.ReverificationSchedule
	var total int64
	query := s.db.WithContext(ctx).Model(&gormmodel.ReverificationSchedule{}).Where("developer_id = ?", developerID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&schedules).Error
	return schedules, total, err
}

// GetSchedule 读取开发者所属计划。
func (s *IdentityStore) GetSchedule(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.ReverificationSchedule, error) {
	var schedule gormmodel.ReverificationSchedule
	if err := s.db.WithContext(ctx).Where("id = ? AND developer_id = ?", id, developerID).Take(&schedule).Error; err != nil {
		return nil, err
	}
	return &schedule, nil
}

// CancelSchedule 取消计划。
func (s *IdentityStore) CancelSchedule(ctx context.Context, developerID, id uuid.UUID) (int64, error) {
	res := s.db.WithContext(ctx).Model(&gormmodel.ReverificationSchedule{}).
		Where("id = ? AND developer_id = ? AND status = 'active'", id, developerID).
		Updates(map[string]any{"status": "cancelled", "updated_at": time.Now().UTC()})
	return res.RowsAffected, res.Error
}

// ExpiringDocuments 列出到期提醒。
func (s *IdentityStore) ExpiringDocuments(ctx context.Context, developerID uuid.UUID, days, page, limit int) ([]gormmodel.ExpiryAlert, int64, error) {
	var alerts []gormmodel.ExpiryAlert
	var total int64
	until := time.Now().UTC().AddDate(0, 0, days)
	query := s.db.WithContext(ctx).Model(&gormmodel.ExpiryAlert{}).
		Where("developer_id = ? AND expiry_date <= ?", developerID, until)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("expiry_date ASC").Offset((page - 1) * limit).Limit(limit).Find(&alerts).Error
	return alerts, total, err
}
