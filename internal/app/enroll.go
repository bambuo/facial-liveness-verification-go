package app

import (
	"context"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// EnrollResult 是人脸录入结果。
type EnrollResult struct {
	ProfileID    uuid.UUID
	Status       string
	ModelVersion string
}

// EnrollService 编排活体、动作、特征与档案写入。
type EnrollService struct {
	support   *VerificationSupport
	actions   *ActionService
	engine    port.FaceEngine
	encryptor port.Encryptor
	subjects  port.SubjectRepo
	profiles  port.ProfileRepo
	audit     *Auditor
	clock     port.Clock
}

// NewEnrollService 构造录入用例。
func NewEnrollService(support *VerificationSupport, actions *ActionService, engine port.FaceEngine, encryptor port.Encryptor, subjects port.SubjectRepo, profiles port.ProfileRepo, audit *Auditor, clock port.Clock) *EnrollService {
	return &EnrollService{support: support, actions: actions, engine: engine, encryptor: encryptor, subjects: subjects, profiles: profiles, audit: audit, clock: clock}
}

// Enroll 执行活体投票、动作校验、特征提取与加密落库。
func (s *EnrollService) Enroll(ctx context.Context, auth *AuthContext, challenge *domain.Challenge, session *domain.Session, frames [][]byte, requestID *string) (EnrollResult, error) {
	if len(frames) == 0 {
		return EnrollResult{}, domain.New(domain.CodeNoFace, "未收到任何帧")
	}
	vote, err := s.engine.CheckLiveness(frames, auth.Policy.LivenessThreshold, auth.Policy.LiveRatioThreshold)
	if err != nil {
		return EnrollResult{}, err
	}
	s.support.RecordLiveness(ctx, auth, challenge.DeviceID, vote.Status)
	liveness := vote.Status
	if vote.Status == domain.LivenessSpoof {
		if _, err := s.support.Finalize(ctx, auth, session, "failed", &liveness, nil, &vote, requestID); err != nil {
			return EnrollResult{}, err
		}
		return EnrollResult{}, domain.Of(domain.CodeLivenessFailed)
	}
	if vote.Status == domain.LivenessFuzzy {
		if _, err := s.support.Finalize(ctx, auth, session, "failed", &liveness, nil, &vote, requestID); err != nil {
			return EnrollResult{}, err
		}
		return EnrollResult{}, domain.Of(domain.CodeLowQuality)
	}
	if err := s.actions.Verify(ctx, challenge.Actions, challenge.FramesPerAction, frames, auth.Policy); err != nil {
		if _, ferr := s.support.Finalize(ctx, auth, session, "failed", &liveness, nil, &vote, requestID); ferr != nil {
			return EnrollResult{}, ferr
		}
		return EnrollResult{}, err
	}
	embedding, err := s.engine.ExtractLargest(frames)
	if err != nil {
		return EnrollResult{}, err
	}
	cipherText, err := s.encryptor.Encrypt(ctx, auth.Tenant.ID, FloatsToBytes(embedding))
	if err != nil {
		return EnrollResult{}, err
	}
	subject, err := s.subjects.GetOrCreate(ctx, auth.Tenant.ID, auth.Claims.Subject)
	if err != nil {
		return EnrollResult{}, domain.New(domain.CodeServiceUnavailable, "主体查询失败")
	}
	now := s.clock.Now().UTC()
	if existing, err := s.profiles.GetActive(ctx, auth.Tenant.ID, subject.ID); err != nil {
		return EnrollResult{}, domain.New(domain.CodeServiceUnavailable, "档案查询失败")
	} else if existing != nil {
		existing.Status = domain.ProfileDeleted
		existing.UpdatedAt = now
		if err := s.profiles.Update(ctx, existing); err != nil {
			return EnrollResult{}, domain.New(domain.CodeServiceUnavailable, "档案更新失败")
		}
	}
	profileID := uuid.New()
	profile := &domain.Profile{
		ID:              profileID,
		TenantID:        auth.Tenant.ID,
		SubjectID:       subject.ID,
		EmbeddingCipher: cipherText,
		KeyVersion:      1,
		ModelVersion:    s.engine.ModelVersion(),
		Status:          domain.ProfileActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if _, err := s.profiles.Upsert(ctx, profile); err != nil {
		return EnrollResult{}, domain.New(domain.CodeServiceUnavailable, "档案写入失败")
	}
	if _, err := s.support.Finalize(ctx, auth, session, "passed", &liveness, nil, &vote, requestID); err != nil {
		return EnrollResult{}, err
	}
	if err := s.audit.Write(ctx, auth.Tenant.ID, auth.Claims.Subject, "enroll.completed", "profile:"+profileID.String(), "ok", requestID, map[string]any{
		"keyVersion": 1,
		"frames":     len(frames),
	}); err != nil {
		return EnrollResult{}, domain.New(domain.CodeServiceUnavailable, "审计写入失败")
	}
	return EnrollResult{ProfileID: profileID, Status: "active", ModelVersion: profile.ModelVersion}, nil
}
