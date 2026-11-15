package app

import (
	"context"
	"fmt"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// VerifyResult 是档案 1:1 的结果。
type VerifyResult struct {
	Result   string
	Score    float64
	Liveness string
	Token    string
}

// VerifyService 编排活体、动作、特征比对与结果令牌。
type VerifyService struct {
	support   *VerificationSupport
	actions   *ActionService
	engine    port.FaceEngine
	encryptor port.Encryptor
	subjects  port.SubjectRepo
	profiles  port.ProfileRepo
}

// NewVerifyService 构造档案核验用例。
func NewVerifyService(support *VerificationSupport, actions *ActionService, engine port.FaceEngine, encryptor port.Encryptor, subjects port.SubjectRepo, profiles port.ProfileRepo) *VerifyService {
	return &VerifyService{support: support, actions: actions, engine: engine, encryptor: encryptor, subjects: subjects, profiles: profiles}
}

// Verify 执行档案 1:1，通过时签发结果令牌。
func (s *VerifyService) Verify(ctx context.Context, auth *AuthContext, challenge *domain.Challenge, session *domain.Session, frames [][]byte, requestID *string) (VerifyResult, error) {
	if len(frames) == 0 {
		return VerifyResult{}, domain.New(domain.CodeNoFace, "未收到任何帧")
	}
	vote, err := s.engine.CheckLiveness(frames, auth.Policy.LivenessThreshold, auth.Policy.LiveRatioThreshold)
	if err != nil {
		return VerifyResult{}, err
	}
	s.support.RecordLiveness(ctx, auth, challenge.DeviceID, vote.Status)
	liveness := vote.Status
	if vote.Status == domain.LivenessSpoof {
		if _, err := s.support.Finalize(ctx, auth, session, "failed", &liveness, nil, &vote, requestID); err != nil {
			return VerifyResult{}, err
		}
		return VerifyResult{}, domain.Of(domain.CodeLivenessFailed)
	}
	if vote.Status == domain.LivenessFuzzy {
		if _, err := s.support.Finalize(ctx, auth, session, "failed", &liveness, nil, &vote, requestID); err != nil {
			return VerifyResult{}, err
		}
		return VerifyResult{}, domain.Of(domain.CodeLowQuality)
	}
	subject, err := s.subjects.GetByExternal(ctx, auth.Tenant.ID, auth.Claims.Subject)
	if err != nil {
		return VerifyResult{}, domain.New(domain.CodeServiceUnavailable, "主体查询失败")
	}
	if subject == nil || subject.Status != domain.SubjectActive {
		return VerifyResult{}, domain.Of(domain.CodeFaceNotEnrolled)
	}
	if err := s.actions.Verify(ctx, challenge.Actions, challenge.FramesPerAction, frames, auth.Policy); err != nil {
		if _, ferr := s.support.Finalize(ctx, auth, session, "failed", &liveness, nil, &vote, requestID); ferr != nil {
			return VerifyResult{}, ferr
		}
		return VerifyResult{}, err
	}
	profile, err := s.profiles.GetActive(ctx, auth.Tenant.ID, subject.ID)
	if err != nil {
		return VerifyResult{}, domain.New(domain.CodeServiceUnavailable, "档案查询失败")
	}
	if profile == nil {
		return VerifyResult{}, domain.Of(domain.CodeFaceNotEnrolled)
	}
	if profile.ModelVersion != s.engine.ModelVersion() {
		return VerifyResult{}, domain.New(domain.CodeServiceUnavailable, "特征版本不一致，请重新录入")
	}
	probe, err := s.engine.ExtractLargest(frames)
	if err != nil {
		return VerifyResult{}, err
	}
	storedBytes, err := s.encryptor.Decrypt(ctx, auth.Tenant.ID, profile.EmbeddingCipher)
	if err != nil {
		return VerifyResult{}, err
	}
	score, err := s.engine.Compare(probe, BytesToFloats(storedBytes))
	if err != nil {
		return VerifyResult{}, err
	}
	rounded := domain.RoundAway(score, 5)
	passed := score >= auth.Policy.MatchThreshold
	result := "failed"
	if passed {
		result = "passed"
	}
	if _, err := s.support.Finalize(ctx, auth, session, result, &liveness, &rounded, &vote, requestID); err != nil {
		return VerifyResult{}, err
	}
	if !passed {
		return VerifyResult{}, domain.New(domain.CodeFaceMismatch, fmt.Sprintf("相似度 %.4f 低于阈值 %g", score, auth.Policy.MatchThreshold))
	}
	token, err := s.support.IssueToken(ctx, auth, challenge.Scene, challenge.Ref, &rounded)
	if err != nil {
		return VerifyResult{}, err
	}
	return VerifyResult{Result: "passed", Score: rounded, Liveness: "real", Token: token}, nil
}
