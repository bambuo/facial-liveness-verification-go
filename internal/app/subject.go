package app

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// SubjectService 提供档案状态查询与合规删除。
type SubjectService struct {
	subjects port.SubjectRepo
	profiles port.ProfileRepo
	storage  port.ObjectStorage
	audits   port.AuditRepo
	clock    port.Clock
}

// NewSubjectService 构造主体用例。
func NewSubjectService(subjects port.SubjectRepo, profiles port.ProfileRepo, storage port.ObjectStorage, audits port.AuditRepo, clock port.Clock) *SubjectService {
	return &SubjectService{subjects: subjects, profiles: profiles, storage: storage, audits: audits, clock: clock}
}

// Get 返回主体状态与 active 档案模型版本。
func (s *SubjectService) Get(ctx context.Context, tenantID uuid.UUID, externalID string) (*domain.Subject, *string, error) {
	subject, err := s.subjects.GetByExternal(ctx, tenantID, externalID)
	if err != nil {
		return nil, nil, domain.New(domain.CodeServiceUnavailable, "主体查询失败")
	}
	if subject == nil || subject.Status == domain.SubjectDeleted {
		return nil, nil, domain.Of(domain.CodeFaceNotEnrolled)
	}
	profile, err := s.profiles.GetActive(ctx, tenantID, subject.ID)
	if err != nil {
		return nil, nil, domain.New(domain.CodeServiceUnavailable, "档案查询失败")
	}
	if profile == nil {
		return subject, nil, nil
	}
	return subject, &profile.ModelVersion, nil
}

// Delete 清理档案、原图与可删除审计字段。
func (s *SubjectService) Delete(ctx context.Context, tenantID uuid.UUID, externalID, actor string, requestID *string) (uuid.UUID, map[string]bool, error) {
	subject, err := s.subjects.GetByExternal(ctx, tenantID, externalID)
	if err != nil {
		return uuid.Nil, nil, domain.New(domain.CodeServiceUnavailable, "主体查询失败")
	}
	if subject == nil {
		return uuid.Nil, nil, domain.Of(domain.CodeFaceNotEnrolled)
	}
	cleaned := map[string]bool{}

	profile, err := s.profiles.GetActive(ctx, tenantID, subject.ID)
	if err != nil {
		return uuid.Nil, nil, domain.New(domain.CodeServiceUnavailable, "档案查询失败")
	}
	if profile != nil {
		now := s.clock.Now().UTC()
		profile.Status = domain.ProfileDeleted
		profile.UpdatedAt = now
		if err := s.profiles.Update(ctx, profile); err != nil {
			return uuid.Nil, nil, domain.New(domain.CodeServiceUnavailable, "档案更新失败")
		}
		cleaned["profile"] = true
	} else {
		cleaned["profile"] = false
	}

	logs, err := s.audits.ListByResource(ctx, tenantID, "subject:"+subject.ID.String())
	if err != nil {
		return uuid.Nil, nil, domain.New(domain.CodeServiceUnavailable, "审计查询失败")
	}
	var keys []string
	for _, log := range logs {
		var detail map[string]any
		if json.Unmarshal(log.DetailJSON, &detail) == nil {
			if raw, ok := detail["keys"].([]any); ok {
				for _, item := range raw {
					if key, ok := item.(string); ok && key != "" {
						keys = append(keys, key)
					}
				}
			}
		}
		log.DetailJSON = []byte("{}")
		if err := s.audits.Update(ctx, log); err != nil {
			return uuid.Nil, nil, domain.New(domain.CodeServiceUnavailable, "审计更新失败")
		}
	}
	for _, key := range keys {
		if err := s.storage.Delete(ctx, key); err != nil {
			slog.Warn("原图删除失败", "key", key, "err", err)
		}
	}
	cleaned["images"] = len(keys) > 0

	now := s.clock.Now().UTC()
	subject.Status = domain.SubjectDeleted
	subject.DeletedAt = &now
	subject.UpdatedAt = now
	if err := s.subjects.Update(ctx, subject); err != nil {
		return uuid.Nil, nil, domain.New(domain.CodeServiceUnavailable, "主体更新失败")
	}
	if err := s.audits.Add(ctx, &domain.AuditLog{
		TenantID:   tenantID,
		Actor:      actor,
		Action:     "subject.deleted",
		Resource:   "subject:" + subject.ID.String(),
		Result:     "deleted",
		RequestID:  requestID,
		DetailJSON: mustJSON(cleaned),
		CreatedAt:  now,
	}); err != nil {
		return uuid.Nil, nil, domain.New(domain.CodeServiceUnavailable, "审计写入失败")
	}
	return subject.ID, cleaned, nil
}

func mustJSON(v any) []byte {
	raw, err := json.Marshal(v)
	if err != nil {
		return []byte("{}")
	}
	return raw
}
