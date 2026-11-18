package pg

import (
	"context"
	"time"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/adapter/pg/gormmodel"
)

// CreateRuleset 创建规则集。
func (s *IdentityStore) CreateRuleset(ctx context.Context, ruleset *gormmodel.ComplianceRuleset) error {
	return s.db.WithContext(ctx).Create(ruleset).Error
}

// ListRulesets 列出开发者规则集。
func (s *IdentityStore) ListRulesets(ctx context.Context, developerID uuid.UUID) ([]gormmodel.ComplianceRuleset, error) {
	var rulesets []gormmodel.ComplianceRuleset
	err := s.db.WithContext(ctx).Where("developer_id = ?", developerID).Order("priority ASC").Find(&rulesets).Error
	return rulesets, err
}

// GetRuleset 读取开发者规则集。
func (s *IdentityStore) GetRuleset(ctx context.Context, developerID, id uuid.UUID) (*gormmodel.ComplianceRuleset, error) {
	var ruleset gormmodel.ComplianceRuleset
	if err := s.db.WithContext(ctx).Where("id = ? AND developer_id = ?", id, developerID).Take(&ruleset).Error; err != nil {
		return nil, err
	}
	return &ruleset, nil
}

// UpdateRuleset 更新规则集。
func (s *IdentityStore) UpdateRuleset(ctx context.Context, developerID, id uuid.UUID, updates map[string]any) (int64, error) {
	updates["updated_at"] = time.Now().UTC()
	res := s.db.WithContext(ctx).Model(&gormmodel.ComplianceRuleset{}).Where("id = ? AND developer_id = ?", id, developerID).Updates(updates)
	return res.RowsAffected, res.Error
}

// DeleteRuleset 删除规则集。
func (s *IdentityStore) DeleteRuleset(ctx context.Context, developerID, id uuid.UUID) (int64, error) {
	res := s.db.WithContext(ctx).Where("id = ? AND developer_id = ?", id, developerID).Delete(&gormmodel.ComplianceRuleset{})
	return res.RowsAffected, res.Error
}

// RulesByRuleset 列出规则集规则。
func (s *IdentityStore) RulesByRuleset(ctx context.Context, rulesetID uuid.UUID) ([]gormmodel.ComplianceRule, error) {
	var rules []gormmodel.ComplianceRule
	err := s.db.WithContext(ctx).Where("ruleset_id = ?", rulesetID).Order("created_at ASC").Find(&rules).Error
	return rules, err
}

// CreateRule 创建规则。
func (s *IdentityStore) CreateRule(ctx context.Context, rule *gormmodel.ComplianceRule) error {
	return s.db.WithContext(ctx).Create(rule).Error
}

// GetRule 读取规则。
func (s *IdentityStore) GetRule(ctx context.Context, id uuid.UUID) (*gormmodel.ComplianceRule, error) {
	var rule gormmodel.ComplianceRule
	if err := s.db.WithContext(ctx).Where("id = ?", id).Take(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// UpdateRule 更新规则。
func (s *IdentityStore) UpdateRule(ctx context.Context, id uuid.UUID, updates map[string]any) (int64, error) {
	res := s.db.WithContext(ctx).Model(&gormmodel.ComplianceRule{}).Where("id = ?", id).Updates(updates)
	return res.RowsAffected, res.Error
}

// DeleteRule 删除规则。
func (s *IdentityStore) DeleteRule(ctx context.Context, id uuid.UUID) (int64, error) {
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&gormmodel.ComplianceRule{})
	return res.RowsAffected, res.Error
}
