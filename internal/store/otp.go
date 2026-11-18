// Package store 提供应用层的原子数据操作（GORM 实现）。
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// OTPResult 是原子 OTP 校验结果。
type OTPResult struct {
	Status       string
	AttemptsLeft int
}

// CheckOTPRateLimit 等价 check_otp_rate_limit：按邮箱窗口计数并返回是否仍可发送。
func CheckOTPRateLimit(ctx context.Context, db *gorm.DB, email string, maxSends, windowSeconds int) (bool, error) {
	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(windowSeconds) * time.Second)
	rec := gormmodel.DeveloperOTPRateLimit{
		Email:       email,
		SendCount:   1,
		WindowStart: now,
		UpdatedAt:   now,
	}
	err := db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "email"}},
			DoUpdates: clause.Assignments(map[string]any{
				"send_count": gorm.Expr(
					"CASE WHEN developer_otp_rate_limits.window_start < ? THEN 1 ELSE developer_otp_rate_limits.send_count + 1 END",
					cutoff,
				),
				"window_start": gorm.Expr(
					"CASE WHEN developer_otp_rate_limits.window_start < ? THEN ? ELSE developer_otp_rate_limits.window_start END",
					cutoff, now,
				),
				"updated_at": now,
			}),
		},
		clause.Returning{},
	).Create(&rec).Error
	if err != nil {
		return false, fmt.Errorf("check otp rate limit: %w", err)
	}
	return int(rec.SendCount) <= maxSends, nil
}

// CheckPhoneOTPRateLimit 等价 check_phone_otp_rate_limit：按验证请求窗口计数。
func CheckPhoneOTPRateLimit(ctx context.Context, db *gorm.DB, verificationID uuid.UUID, maxSends, windowSeconds int) (bool, error) {
	now := time.Now().UTC()
	cutoff := now.Add(-time.Duration(windowSeconds) * time.Second)
	rec := gormmodel.PhoneOTPRateLimit{
		VerificationRequestID: verificationID,
		SendCount:             1,
		WindowStart:           now,
		UpdatedAt:             now,
	}
	err := db.WithContext(ctx).Clauses(
		clause.OnConflict{
			Columns: []clause.Column{{Name: "verification_request_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"send_count": gorm.Expr(
					"CASE WHEN phone_otp_rate_limits.window_start < ? THEN 1 ELSE phone_otp_rate_limits.send_count + 1 END",
					cutoff,
				),
				"window_start": gorm.Expr(
					"CASE WHEN phone_otp_rate_limits.window_start < ? THEN ? ELSE phone_otp_rate_limits.window_start END",
					cutoff, now,
				),
				"updated_at": now,
			}),
		},
		clause.Returning{},
	).Create(&rec).Error
	if err != nil {
		return false, fmt.Errorf("check phone otp rate limit: %w", err)
	}
	return int(rec.SendCount) <= maxSends, nil
}

// VerifyOTPAtomic 等价 verify_otp_atomic：锁最新未用验证码并原子消费。
func VerifyOTPAtomic(ctx context.Context, tx *gorm.DB, email, codeHash string, maxAttempts int) (OTPResult, error) {
	var row gormmodel.DeveloperOTPCode
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("email = ? AND used_at IS NULL AND expires_at > ?", email, time.Now().UTC()).
		Order("created_at DESC").
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return OTPResult{Status: "not_found"}, nil
	}
	if err != nil {
		return OTPResult{}, fmt.Errorf("select otp: %w", err)
	}
	attempts := int(row.Attempts)
	if attempts >= maxAttempts {
		if err := tx.WithContext(ctx).Model(&gormmodel.DeveloperOTPCode{}).
			Where("id = ?", row.ID).
			UpdateColumn("used_at", time.Now().UTC()).Error; err != nil {
			return OTPResult{}, err
		}
		return OTPResult{Status: "exhausted"}, nil
	}
	if err := tx.WithContext(ctx).Model(&gormmodel.DeveloperOTPCode{}).
		Where("id = ?", row.ID).
		UpdateColumn("attempts", attempts+1).Error; err != nil {
		return OTPResult{}, err
	}
	left := maxAttempts - attempts - 1
	if row.CodeHash == codeHash {
		if err := tx.WithContext(ctx).Model(&gormmodel.DeveloperOTPCode{}).
			Where("id = ?", row.ID).
			UpdateColumn("used_at", time.Now().UTC()).Error; err != nil {
			return OTPResult{}, err
		}
		return OTPResult{Status: "valid", AttemptsLeft: left}, nil
	}
	if left <= 0 {
		if err := tx.WithContext(ctx).Model(&gormmodel.DeveloperOTPCode{}).
			Where("id = ?", row.ID).
			UpdateColumn("used_at", time.Now().UTC()).Error; err != nil {
			return OTPResult{}, err
		}
	}
	if left < 0 {
		left = 0
	}
	return OTPResult{Status: "invalid", AttemptsLeft: left}, nil
}

// VerifyPhoneOTPAtomic 等价 verify_phone_otp_atomic；exhausted/末次失败不标记 used。
func VerifyPhoneOTPAtomic(ctx context.Context, tx *gorm.DB, verificationID uuid.UUID, codeHash string, maxAttempts int) (OTPResult, error) {
	var row gormmodel.PhoneOTPCode
	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("verification_request_id = ? AND used_at IS NULL AND expires_at > ?", verificationID, time.Now().UTC()).
		Order("created_at DESC").
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return OTPResult{Status: "not_found"}, nil
	}
	if err != nil {
		return OTPResult{}, fmt.Errorf("select phone otp: %w", err)
	}
	attempts := int(row.Attempts)
	if attempts >= maxAttempts {
		return OTPResult{Status: "exhausted"}, nil
	}
	if err := tx.WithContext(ctx).Model(&gormmodel.PhoneOTPCode{}).
		Where("id = ?", row.ID).
		UpdateColumn("attempts", attempts+1).Error; err != nil {
		return OTPResult{}, err
	}
	left := maxAttempts - attempts - 1
	if row.CodeHash == codeHash {
		if err := tx.WithContext(ctx).Model(&gormmodel.PhoneOTPCode{}).
			Where("id = ?", row.ID).
			UpdateColumn("used_at", time.Now().UTC()).Error; err != nil {
			return OTPResult{}, err
		}
		return OTPResult{Status: "valid", AttemptsLeft: left}, nil
	}
	return OTPResult{Status: "invalid", AttemptsLeft: left}, nil
}

// CleanupOldRateLimits 等价 cleanup_old_rate_limits：删除 48 小时前的未锁定窗口。
func CleanupOldRateLimits(ctx context.Context, db *gorm.DB) (int64, error) {
	res := db.WithContext(ctx).
		Where("window_start < ?", time.Now().UTC().Add(-48*time.Hour)).
		Where("blocked_until IS NULL").
		Delete(&gormmodel.RateLimit{})
	if res.Error != nil {
		return 0, fmt.Errorf("cleanup rate limits: %w", res.Error)
	}
	return res.RowsAffected, nil
}
