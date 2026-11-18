package pg

import (
	"context"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// GetActiveCredential 读取验证请求的未撤销凭证。
func (s *IdentityStore) GetActiveCredential(ctx context.Context, developerID, verificationID uuid.UUID) (*gormmodel.VerifiableCredential, error) {
	var credential gormmodel.VerifiableCredential
	if err := s.db.WithContext(ctx).
		Where("developer_id = ? AND verification_request_id = ? AND revoked_at IS NULL", developerID, verificationID).
		Take(&credential).Error; err != nil {
		return nil, err
	}
	return &credential, nil
}

// CreateCredential 保存签发的凭证。
func (s *IdentityStore) CreateCredential(ctx context.Context, credential *gormmodel.VerifiableCredential) error {
	return s.db.WithContext(ctx).Create(credential).Error
}

// GetCredentialByJTI 按 JTI 读取凭证。
func (s *IdentityStore) GetCredentialByJTI(ctx context.Context, jti string) (*gormmodel.VerifiableCredential, error) {
	var credential gormmodel.VerifiableCredential
	if err := s.db.WithContext(ctx).Where("credential_jti = ?", jti).Take(&credential).Error; err != nil {
		return nil, err
	}
	return &credential, nil
}

// RevokeCredential 撤销开发者所属凭证。
func (s *IdentityStore) RevokeCredential(ctx context.Context, developerID uuid.UUID, jti, reason string) (int64, error) {
	now := time.Now().UTC()
	updates := map[string]any{"revoked_at": now}
	if reason != "" {
		updates["revocation_reason"] = reason
	}
	res := s.db.WithContext(ctx).Model(&gormmodel.VerifiableCredential{}).
		Where("developer_id = ? AND credential_jti = ? AND revoked_at IS NULL", developerID, jti).
		Updates(updates)
	return res.RowsAffected, res.Error
}
