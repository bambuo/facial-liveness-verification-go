package pg

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
	"facial-liveness-verification-go/internal/port"
	"facial-liveness-verification-go/internal/store"
)

// IdentityStore 提供 idswyft API Key、handoff/session token 与核心验证记录读写。
type IdentityStore struct {
	db *gorm.DB
}

// NewIdentityStore 构造 idswyft 鉴权仓储。
func NewIdentityStore(db *gorm.DB) *IdentityStore {
	return &IdentityStore{db: db}
}

// FindAPIKeyByHash 按 HMAC 哈希查找有效 API Key 及其开发者状态。
func (s *IdentityStore) FindAPIKeyByHash(ctx context.Context, hash string) (*port.APIKeyAuth, error) {
	var key gormmodel.APIKey
	if err := s.db.WithContext(ctx).
		Where("key_hash = ? AND is_active = TRUE", hash).
		Take(&key).Error; err != nil {
		return nil, err
	}
	out := &port.APIKeyAuth{
		ID:        key.ID,
		IsSandbox: key.IsSandbox != nil && *key.IsSandbox,
		IsService: key.IsService,
		ExpiresAt: key.ExpiresAt,
	}
	if key.DeveloperID != nil {
		out.DeveloperID = *key.DeveloperID
		var developer gormmodel.Developer
		if err := s.db.WithContext(ctx).Where("id = ?", *key.DeveloperID).Take(&developer).Error; err != nil {
			return nil, err
		}
		out.DeveloperStatus = developer.Status
	}
	return out, nil
}

// FindHandoffByToken 按 token 哈希查找 handoff 会话并解析关联 API Key。
func (s *IdentityStore) FindHandoffByToken(ctx context.Context, hash string) (*port.HandoffAuth, error) {
	var session gormmodel.MobileHandoffSession
	if err := s.db.WithContext(ctx).Where("token = ?", hash).Take(&session).Error; err != nil {
		return nil, err
	}
	key, err := s.findAPIKeyByID(ctx, session.APIKeyID)
	if err != nil {
		return nil, err
	}
	return &port.HandoffAuth{
		APIKey:         *key,
		ExpiresAt:      session.ExpiresAt,
		Status:         session.Status,
		VerificationID: session.VerificationID,
	}, nil
}

// FindSessionByToken 按 session token 哈希查找验证请求。
func (s *IdentityStore) FindSessionByToken(ctx context.Context, hash string) (*port.SessionAuth, error) {
	var verification gormmodel.VerificationRequest
	if err := s.db.WithContext(ctx).Where("session_token_hash = ?", hash).Take(&verification).Error; err != nil {
		return nil, err
	}
	out := &port.SessionAuth{
		VerificationID: verification.ID,
		ExpiresAt:      verification.SessionTokenExpiresAt,
		APIKeyID:       verification.SessionAPIKeyID,
	}
	if verification.DeveloperID != nil {
		out.DeveloperID = *verification.DeveloperID
	}
	return out, nil
}

// TouchAPIKey 更新 API Key 最后使用时间。
func (s *IdentityStore) TouchAPIKey(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&gormmodel.APIKey{}).
		Where("id = ?", id).Update("last_used_at", time.Now().UTC()).Error
}

func (s *IdentityStore) findAPIKeyByID(ctx context.Context, id uuid.UUID) (*port.APIKeyAuth, error) {
	var key gormmodel.APIKey
	if err := s.db.WithContext(ctx).Where("id = ? AND is_active = TRUE", id).Take(&key).Error; err != nil {
		return nil, err
	}
	out := &port.APIKeyAuth{
		ID: key.ID, IsSandbox: key.IsSandbox != nil && *key.IsSandbox,
		IsService: key.IsService, ExpiresAt: key.ExpiresAt,
	}
	if key.DeveloperID != nil {
		out.DeveloperID = *key.DeveloperID
		var developer gormmodel.Developer
		if err := s.db.WithContext(ctx).Where("id = ?", *key.DeveloperID).Take(&developer).Error; err != nil {
			return nil, err
		}
		out.DeveloperStatus = developer.Status
	}
	return out, nil
}

// CreateVerification 插入验证请求。
func (s *IdentityStore) CreateVerification(ctx context.Context, verification *gormmodel.VerificationRequest) error {
	return s.db.WithContext(ctx).Create(verification).Error
}

// GetVerification 读取验证请求；不存在时返回 gorm.ErrRecordNotFound。
func (s *IdentityStore) GetVerification(ctx context.Context, id uuid.UUID) (*gormmodel.VerificationRequest, error) {
	var verification gormmodel.VerificationRequest
	if err := s.db.WithContext(ctx).Where("id = ?", id).Take(&verification).Error; err != nil {
		return nil, err
	}
	return &verification, nil
}

// UpdateVerification 按字段 map 更新验证请求。
func (s *IdentityStore) UpdateVerification(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return s.db.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).Where("id = ?", id).Updates(updates).Error
}

// CreateDocument 插入证件文档。
func (s *IdentityStore) CreateDocument(ctx context.Context, document *gormmodel.Document) error {
	return s.db.WithContext(ctx).Create(document).Error
}

// GetDocument 读取证件文档。
func (s *IdentityStore) GetDocument(ctx context.Context, id uuid.UUID) (*gormmodel.Document, error) {
	var document gormmodel.Document
	if err := s.db.WithContext(ctx).Where("id = ?", id).Take(&document).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

// GetBackDocument 读取验证请求最近的后置证件。
func (s *IdentityStore) GetBackDocument(ctx context.Context, verificationID uuid.UUID) (*gormmodel.Document, error) {
	var document gormmodel.Document
	if err := s.db.WithContext(ctx).
		Where("verification_request_id = ? AND is_back_of_id = TRUE", verificationID).
		Order("created_at DESC").Take(&document).Error; err != nil {
		return nil, err
	}
	return &document, nil
}

// CreateSelfie 插入自拍记录。
func (s *IdentityStore) CreateSelfie(ctx context.Context, selfie *gormmodel.Selfie) error {
	return s.db.WithContext(ctx).Create(selfie).Error
}

// GetSelfie 读取自拍记录。
func (s *IdentityStore) GetSelfie(ctx context.Context, id uuid.UUID) (*gormmodel.Selfie, error) {
	var selfie gormmodel.Selfie
	if err := s.db.WithContext(ctx).Where("id = ?", id).Take(&selfie).Error; err != nil {
		return nil, err
	}
	return &selfie, nil
}

// UpsertContext 写入验证上下文 JSON。
func (s *IdentityStore) UpsertContext(ctx context.Context, verificationID uuid.UUID, payload []byte) error {
	row := gormmodel.VerificationContext{VerificationID: verificationID, Context: payload, UpdatedAt: time.Now().UTC()}
	return s.db.WithContext(ctx).Save(&row).Error
}

// GetContext 读取验证上下文 JSON。
func (s *IdentityStore) GetContext(ctx context.Context, verificationID uuid.UUID) ([]byte, error) {
	var row gormmodel.VerificationContext
	if err := s.db.WithContext(ctx).Where("verification_id = ?", verificationID).Take(&row).Error; err != nil {
		return nil, err
	}
	return row.Context, nil
}

// IsNotFound 判断仓储查询是否无记录。
func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// FindDeveloperByEmail 按邮箱查找已注册开发者。
func (s *IdentityStore) FindDeveloperByEmail(ctx context.Context, email string) (*gormmodel.Developer, error) {
	var developer gormmodel.Developer
	if err := s.db.WithContext(ctx).Where("email = ?", email).Take(&developer).Error; err != nil {
		return nil, err
	}
	return &developer, nil
}

// CheckDeveloperOTPSend 检查邮箱 OTP 发送窗口。
func (s *IdentityStore) CheckDeveloperOTPSend(ctx context.Context, email string, maxSends, windowSeconds int) (bool, error) {
	return store.CheckOTPRateLimit(ctx, s.db, email, maxSends, windowSeconds)
}

// CreateDeveloperOTP 写入邮箱 OTP，并作废旧验证码。
func (s *IdentityStore) CreateDeveloperOTP(ctx context.Context, email, codeHash string, ttl time.Duration) error {
	now := time.Now().UTC()
	if err := s.db.WithContext(ctx).Model(&gormmodel.DeveloperOTPCode{}).
		Where("email = ? AND used_at IS NULL", email).Update("used_at", now).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(&gormmodel.DeveloperOTPCode{
		Email: email, CodeHash: codeHash, ExpiresAt: now.Add(ttl), CreatedAt: now,
	}).Error
}

// VerifyDeveloperOTP 原子校验邮箱 OTP。
func (s *IdentityStore) VerifyDeveloperOTP(ctx context.Context, email, codeHash string, maxAttempts int) (store.OTPResult, error) {
	var result store.OTPResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = store.VerifyOTPAtomic(ctx, tx, email, codeHash, maxAttempts)
		return err
	})
	return result, err
}

// WebhookForDeveloper 返回开发者当前启用的首个 Webhook 目标。
func (s *IdentityStore) WebhookForDeveloper(ctx context.Context, developerID uuid.UUID) (*port.WebhookTarget, error) {
	var webhook gormmodel.Webhook
	if err := s.db.WithContext(ctx).
		Where("developer_id = ? AND is_active = TRUE", developerID).
		Order("created_at DESC").Take(&webhook).Error; err != nil {
		return nil, err
	}
	target := &port.WebhookTarget{ID: webhook.ID, URL: webhook.URL}
	if webhook.SecretKey != nil {
		target.Secret = *webhook.SecretKey
	}
	return target, nil
}

// CreateWebhookDelivery 记录一次 Webhook 投递。
func (s *IdentityStore) CreateWebhookDelivery(ctx context.Context, delivery *gormmodel.WebhookDelivery) error {
	return s.db.WithContext(ctx).Create(delivery).Error
}

// UpdateWebhookDelivery 更新 Webhook 投递状态。
func (s *IdentityStore) UpdateWebhookDelivery(ctx context.Context, id uuid.UUID, updates map[string]any) error {
	return s.db.WithContext(ctx).Model(&gormmodel.WebhookDelivery{}).Where("id = ?", id).Updates(updates).Error
}

// CheckPhoneOTPSend 检查验证请求的短信发送窗口。
func (s *IdentityStore) CheckPhoneOTPSend(ctx context.Context, verificationID uuid.UUID, maxSends, windowSeconds int) (bool, error) {
	return store.CheckPhoneOTPRateLimit(ctx, s.db, verificationID, maxSends, windowSeconds)
}

// CreatePhoneOTP 写入带哈希的短信验证码。
func (s *IdentityStore) CreatePhoneOTP(ctx context.Context, verificationID uuid.UUID, phone, codeHash string, ttl time.Duration) error {
	now := time.Now().UTC()
	return s.db.WithContext(ctx).Create(&gormmodel.PhoneOTPCode{
		VerificationRequestID: verificationID,
		PhoneNumber:           phone,
		CodeHash:              codeHash,
		ExpiresAt:             now.Add(ttl),
		CreatedAt:             now,
	}).Error
}

// VerifyPhoneOTP 原子校验短信验证码。
func (s *IdentityStore) VerifyPhoneOTP(ctx context.Context, verificationID uuid.UUID, codeHash string, maxAttempts int) (store.OTPResult, error) {
	var result store.OTPResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = store.VerifyPhoneOTPAtomic(ctx, tx, verificationID, codeHash, maxAttempts)
		return err
	})
	return result, err
}
