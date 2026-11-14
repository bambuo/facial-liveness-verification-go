package app

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"math"

	"github.com/google/uuid"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// Auditor 写入只含非敏感元数据的审计。
type Auditor struct {
	repo  port.AuditRepo
	clock port.Clock
}

// NewAuditor 构造审计写入器。
func NewAuditor(repo port.AuditRepo, clock port.Clock) *Auditor {
	return &Auditor{repo: repo, clock: clock}
}

// Write 写审计；detail 为 nil 时保存空对象。
func (a *Auditor) Write(ctx context.Context, tenantID uuid.UUID, actor, action, resource, result string, requestID *string, detail any) error {
	payload := []byte("{}")
	if detail != nil {
		if raw, err := json.Marshal(detail); err == nil {
			payload = raw
		}
	}
	return a.repo.Add(ctx, &domain.AuditLog{
		TenantID:   tenantID,
		Actor:      actor,
		Action:     action,
		Resource:   resource,
		Result:     result,
		RequestID:  requestID,
		DetailJSON: payload,
		CreatedAt:  a.clock.Now().UTC(),
	})
}

// VerificationSupport 负责 challenge 一次性消费、会话收尾与结果令牌。
type VerificationSupport struct {
	challenges port.ChallengeStore
	sessions   port.SessionRepo
	subjects   port.SubjectRepo
	issuer     port.TokenIssuer
	audit      *Auditor
	clock      port.Clock
}

// NewVerificationSupport 构造上传用例共享逻辑。
func NewVerificationSupport(challenges port.ChallengeStore, sessions port.SessionRepo, subjects port.SubjectRepo, issuer port.TokenIssuer, audit *Auditor, clock port.Clock) *VerificationSupport {
	return &VerificationSupport{challenges: challenges, sessions: sessions, subjects: subjects, issuer: issuer, audit: audit, clock: clock}
}

// BeginUpload 原子消费 challenge、校验绑定并推进会话到 processing。
func (s *VerificationSupport) BeginUpload(ctx context.Context, auth *AuthContext, challengeID, deviceID string, mode domain.MatchMode) (*domain.Challenge, *domain.Session, error) {
	challenge, err := s.challenges.Consume(ctx, challengeID)
	if err != nil {
		return nil, nil, domain.New(domain.CodeServiceUnavailable, "challenge 存储不可用")
	}
	if challenge == nil || challenge.ExpiresAt.Before(s.clock.Now()) {
		return nil, nil, domain.Of(domain.CodeChallengeExpired)
	}
	if challenge.TenantID != auth.Tenant.ID {
		return nil, nil, domain.New(domain.CodeChallengeReplay, "挑战归属校验失败")
	}
	if challenge.ExternalSubject != auth.Claims.Subject {
		return nil, nil, domain.New(domain.CodeChallengeReplay, "挑战归属校验失败")
	}
	if challenge.Mode != mode {
		return nil, nil, domain.New(domain.CodeChallengeReplay, "挑战模式不匹配")
	}
	if challenge.DeviceID != deviceID {
		return nil, nil, domain.New(domain.CodeChallengeReplay, "设备绑定不一致")
	}
	session, err := s.sessions.GetByChallengeID(ctx, auth.Tenant.ID, challengeID)
	if err != nil {
		return nil, nil, domain.New(domain.CodeServiceUnavailable, "会话查询失败")
	}
	if session == nil {
		return nil, nil, domain.New(domain.CodeChallengeExpired, "会话不存在")
	}
	if session.Status == domain.SessionPassed || session.Status == domain.SessionFailed {
		return nil, nil, domain.New(domain.CodeChallengeReplay, "会话已完成")
	}
	subject, err := s.subjects.GetOrCreate(ctx, auth.Tenant.ID, auth.Claims.Subject)
	if err != nil {
		return nil, nil, domain.New(domain.CodeServiceUnavailable, "主体查询失败")
	}
	session.SubjectID = &subject.ID
	session.Status = domain.SessionProcessing
	if err := s.sessions.Update(ctx, session); err != nil {
		return nil, nil, domain.New(domain.CodeServiceUnavailable, "会话更新失败")
	}
	return challenge, session, nil
}

// Finalize 更新会话结果；passed 时签发结果令牌并写审计。
func (s *VerificationSupport) Finalize(ctx context.Context, auth *AuthContext, session *domain.Session, result string, liveness *domain.LivenessStatus, score *float64, vote *domain.LivenessVote, requestID *string) (string, error) {
	if result == "passed" {
		session.Status = domain.SessionPassed
	} else {
		session.Status = domain.SessionFailed
	}
	if score != nil {
		rounded := domain.RoundAway(*score, 5)
		session.Score = &rounded
	}
	session.Liveness = liveness
	if vote != nil {
		clarity := domain.RoundAway(vote.MinScore, 5)
		reality := domain.RoundAway(vote.LiveRatio, 5)
		session.Clarity = &clarity
		session.Reality = &reality
	}
	now := s.clock.Now().UTC()
	session.CompletedAt = &now
	if err := s.sessions.Update(ctx, session); err != nil {
		return "", domain.New(domain.CodeServiceUnavailable, "会话更新失败")
	}

	token := ""
	if result == "passed" {
		issued, err := s.issuer.IssueResult(domain.ResultTokenPayload{
			TenantCode: auth.Tenant.Code,
			Subject:    auth.Claims.Subject,
			Scene:      auth.Claims.Scene,
			Ref:        auth.Claims.Ref,
			Result:     "passed",
			Liveness:   liveness,
			Score:      score,
			Jti:        randomHex(16),
		})
		if err != nil {
			return "", err
		}
		token = issued
	}
	detail := map[string]any{"mode": string(session.Mode), "score": score}
	if liveness != nil {
		detail["liveness"] = string(*liveness)
	}
	if err := s.audit.Write(ctx, auth.Tenant.ID, auth.Claims.Subject, "verification.completed", "session:"+session.ID.String(), result, requestID, detail); err != nil {
		return "", domain.New(domain.CodeServiceUnavailable, "审计写入失败")
	}
	return token, nil
}

// IssueToken 为指定场景签发结果令牌。
func (s *VerificationSupport) IssueToken(ctx context.Context, auth *AuthContext, scene string, ref *string, score *float64) (string, error) {
	liveness := domain.LivenessReal
	return s.issuer.IssueResult(domain.ResultTokenPayload{
		TenantCode: auth.Tenant.Code,
		Subject:    auth.Claims.Subject,
		Scene:      scene,
		Ref:        ref,
		Result:     "passed",
		Liveness:   &liveness,
		Score:      score,
		Jti:        randomHex(16),
	})
}

// FloatsToBytes 将 float32 向量按小端序序列化。
func FloatsToBytes(v []float32) []byte {
	out := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(out[i*4:], math.Float32bits(f))
	}
	return out
}

// BytesToFloats 将小端序字节还原为 float32 向量。
func BytesToFloats(b []byte) []float32 {
	out := make([]float32, len(b)/4)
	for i := range out {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4:]))
	}
	return out
}
