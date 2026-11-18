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

// VerificationUpdate 是状态更新时可选的分数字段。
type VerificationUpdate struct {
	FaceMatchScore       *float64
	LivenessScore        *float64
	CrossValidationScore *float64
}

// UpdateVerificationWithStateCheck 等价 update_verification_with_state_check：
// 锁定行、校验期望状态后原子更新状态与分数。
func UpdateVerificationWithStateCheck(
	ctx context.Context,
	tx *gorm.DB,
	verificationID uuid.UUID,
	expected []string,
	newStatus string,
	update VerificationUpdate,
) (bool, error) {
	var row struct {
		Status string `gorm:"column:status"`
	}
	err := tx.WithContext(ctx).
		Model(&gormmodel.VerificationRequest{}).
		Select("status").
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", verificationID).
		Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("select verification: %w", err)
	}
	matched := false
	for _, state := range expected {
		if row.Status == state {
			matched = true
			break
		}
	}
	if !matched {
		return false, nil
	}
	updates := map[string]any{
		"status":     newStatus,
		"updated_at": time.Now().UTC(),
	}
	if update.FaceMatchScore != nil {
		updates["face_match_score"] = *update.FaceMatchScore
	}
	if update.LivenessScore != nil {
		updates["liveness_score"] = *update.LivenessScore
	}
	if update.CrossValidationScore != nil {
		updates["cross_validation_score"] = *update.CrossValidationScore
	}
	res := tx.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).
		Where("id = ?", verificationID).
		Updates(updates)
	if res.Error != nil {
		return false, fmt.Errorf("update verification: %w", res.Error)
	}
	return res.RowsAffected > 0, nil
}

// TouchUpdatedAt 等价 update_updated_at_column：只允许白名单表。
func TouchUpdatedAt(ctx context.Context, db *gorm.DB, table string, id uuid.UUID) error {
	var model any
	switch table {
	case "users":
		model = &gormmodel.User{}
	case "developers":
		model = &gormmodel.Developer{}
	case "verification_requests":
		model = &gormmodel.VerificationRequest{}
	default:
		return fmt.Errorf("touch updated_at: 不允许的表 %q", table)
	}
	if err := db.WithContext(ctx).Model(model).
		Where("id = ?", id).
		Update("updated_at", time.Now().UTC()).Error; err != nil {
		return fmt.Errorf("touch %s: %w", table, err)
	}
	return nil
}

// TouchVerificationContext 等价 update_verification_contexts_updated_at。
func TouchVerificationContext(ctx context.Context, db *gorm.DB, verificationID uuid.UUID) error {
	err := db.WithContext(ctx).Model(&gormmodel.VerificationContext{}).
		Where("verification_id = ?", verificationID).
		Update("updated_at", time.Now().UTC()).Error
	if err != nil {
		return fmt.Errorf("touch verification_contexts: %w", err)
	}
	return nil
}

// MarkBackOfIDUploaded 等价 update_back_of_id_status 的主更新。
func MarkBackOfIDUploaded(ctx context.Context, tx *gorm.DB, verificationID uuid.UUID) error {
	err := tx.WithContext(ctx).Model(&gormmodel.VerificationRequest{}).
		Where("id = ?", verificationID).
		Updates(map[string]any{
			"back_of_id_uploaded": true,
			"updated_at":          time.Now().UTC(),
		}).Error
	if err != nil {
		return fmt.Errorf("mark back_of_id_uploaded: %w", err)
	}
	return nil
}

// LinkFrontBackDocuments 等价 link_front_and_back_documents：
// 找到同验证请求、同证件类型的最近正面文档并双向关联。
func LinkFrontBackDocuments(ctx context.Context, tx *gorm.DB, backDocID, verificationID uuid.UUID, documentType string) (bool, error) {
	var front struct {
		ID uuid.UUID `gorm:"column:id"`
	}
	err := tx.WithContext(ctx).
		Model(&gormmodel.Document{}).
		Select("id").
		Where("verification_request_id = ? AND is_back_of_id = FALSE AND document_type = ?", verificationID, documentType).
		Order("created_at DESC").
		Take(&front).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("select front document: %w", err)
	}
	if err := tx.WithContext(ctx).Model(&gormmodel.Document{}).
		Where("id = ?", backDocID).
		Update("back_of_id_document_id", front.ID).Error; err != nil {
		return false, fmt.Errorf("link back document: %w", err)
	}
	if err := tx.WithContext(ctx).Model(&gormmodel.Document{}).
		Where("id = ?", front.ID).
		Update("back_of_id_document_id", backDocID).Error; err != nil {
		return false, fmt.Errorf("link front document: %w", err)
	}
	return true, nil
}

// DocumentInput 是插入文档后触发逻辑所需字段。
type DocumentInput struct {
	ID                    uuid.UUID
	VerificationRequestID uuid.UUID
	DocumentType          string
	IsBack                bool
}

// AfterDocumentInsert 等价两个 AFTER INSERT 触发器的组合逻辑。
func AfterDocumentInsert(ctx context.Context, tx *gorm.DB, doc DocumentInput) error {
	if !doc.IsBack {
		return nil
	}
	if err := MarkBackOfIDUploaded(ctx, tx, doc.VerificationRequestID); err != nil {
		return err
	}
	if _, err := LinkFrontBackDocuments(ctx, tx, doc.ID, doc.VerificationRequestID, doc.DocumentType); err != nil {
		return err
	}
	return nil
}
