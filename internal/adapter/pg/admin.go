package pg

import (
	"context"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// AdminVerifications 分页查询验证请求。
func (s *IdentityStore) AdminVerifications(ctx context.Context, status string, developerID *uuid.UUID, page, limit int) ([]gormmodel.VerificationRequest, int64, error) {
	var rows []gormmodel.VerificationRequest
	var total int64
	query := s.db.WithContext(ctx).Model(&gormmodel.VerificationRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if developerID != nil {
		query = query.Where("developer_id = ?", *developerID)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&rows).Error
	return rows, total, err
}

// AdminStatusCounts 返回状态计数。
func (s *IdentityStore) AdminStatusCounts(ctx context.Context, since time.Time) (map[string]int64, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).
		Select("status, count(*) as count").Where("created_at >= ?", since).Group("status").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, item := range rows {
		out[item.Status] = item.Count
	}
	return out, nil
}

// AdminDeveloperCount 返回开发者数量。
func (s *IdentityStore) AdminDeveloperCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&gormmodel.Developer{}).Count(&count).Error
	return count, err
}

// DuplicateVerifications 返回同 user 的其他验证请求。
func (s *IdentityStore) DuplicateVerifications(ctx context.Context, verification *gormmodel.VerificationRequest) ([]gormmodel.VerificationRequest, error) {
	var rows []gormmodel.VerificationRequest
	if verification.UserID == nil {
		return rows, nil
	}
	err := s.db.WithContext(ctx).Where("user_id = ? AND id <> ?", *verification.UserID, verification.ID).Limit(50).Find(&rows).Error
	return rows, err
}

// ReviewVerification 更新审核结果。
func (s *IdentityStore) ReviewVerification(ctx context.Context, id uuid.UUID, status, reason, reviewer string) error {
	now := time.Now().UTC()
	updates := map[string]any{"status": status, "reviewed_at": now, "reviewed_by": reviewer, "updated_at": now}
	if reason != "" {
		updates["manual_review_reason"] = reason
	}
	return s.db.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).Where("id = ?", id).Updates(updates).Error
}

// ProviderMetricSummary 返回供应商指标聚合。
func (s *IdentityStore) ProviderMetricSummary(ctx context.Context, since time.Time) ([]map[string]any, error) {
	type row struct {
		ProviderName string
		Total        int64
		Successes    int64
		AvgLatency   float64
	}
	var rows []row
	err := s.db.WithContext(ctx).Model(&gormmodel.ProviderMetric{}).
		Select("provider_name, count(*) as total, sum(case when success then 1 else 0 end) as successes, coalesce(avg(latency_ms),0) as avg_latency").
		Where("created_at >= ?", since).Group("provider_name").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{"provider_name": r.ProviderName, "total": r.Total, "successes": r.Successes, "avg_latency_ms": r.AvgLatency})
	}
	return out, nil
}

// AuditLogs 导出活动日志。
func (s *IdentityStore) AuditLogs(ctx context.Context, from, to time.Time, developerID *uuid.UUID, status *int32) ([]gormmodel.APIActivityLog, error) {
	var logs []gormmodel.APIActivityLog
	query := s.db.WithContext(ctx).Where("timestamp >= ? AND timestamp <= ?", from, to)
	if developerID != nil {
		query = query.Where("developer_id = ?", *developerID)
	}
	if status != nil {
		query = query.Where("status_code = ?", *status)
	}
	err := query.Order("timestamp DESC").Limit(10000).Find(&logs).Error
	return logs, err
}

// DeleteUserData 删除用户及其验证记录（外键级联）。
func (s *IdentityStore) DeleteUserData(ctx context.Context, userID uuid.UUID) (int64, error) {
	res := s.db.WithContext(ctx).Where("id = ?", userID).Delete(&gormmodel.User{})
	return res.RowsAffected, res.Error
}
