package app

import (
	"context"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// LivenessResult 是单次活体检查的完整证据。
type LivenessResult struct {
	Result        string
	Status        domain.LivenessStatus
	MinScore      float64
	LiveRatio     float64
	FrameScores   []float64
	LiveFrames    int
	AttackFrames  int
	ActionChecked bool
	ActionPassed  bool
	ActionCode    string
	ActionDetail  string
}

// LivenessService 执行仅活体检查，供 lab 测试接口使用。
type LivenessService struct {
	support *VerificationSupport
	actions *ActionService
	engine  port.FaceEngine
}

// NewLivenessService 构造活体检查用例。
func NewLivenessService(support *VerificationSupport, actions *ActionService, engine port.FaceEngine) *LivenessService {
	return &LivenessService{support: support, actions: actions, engine: engine}
}

// Check 执行活体投票并收尾会话；spoof/fuzzy 作为结果返回，不抛业务错误。
func (s *LivenessService) Check(ctx context.Context, auth *AuthContext, challenge *domain.Challenge, session *domain.Session, frames [][]byte, verifyActions bool, requestID *string) (LivenessResult, error) {
	if len(frames) == 0 {
		return LivenessResult{}, domain.New(domain.CodeNoFace, "未收到任何帧")
	}
	vote, err := s.engine.CheckLiveness(frames, auth.Policy.LivenessThreshold, auth.Policy.LiveRatioThreshold)
	if err != nil {
		return LivenessResult{}, err
	}
	s.support.RecordLiveness(ctx, auth, challenge.DeviceID, vote.Status)
	actionChecked := verifyActions
	actionPassed := true
	actionCode, actionDetail := "", ""
	if actionChecked {
		if err := s.actions.Verify(ctx, challenge.Actions, challenge.FramesPerAction, frames, auth.Policy); err != nil {
			actionPassed = false
			if de, ok := err.(*domain.Error); ok {
				actionCode = string(de.Code)
				actionDetail = de.Message
			} else {
				actionDetail = err.Error()
			}
		}
	}
	result := "failed"
	if vote.Status == domain.LivenessReal && (!actionChecked || actionPassed) {
		result = "passed"
	}
	liveness := vote.Status
	if _, err := s.support.Finalize(ctx, auth, session, result, &liveness, nil, &vote, requestID); err != nil {
		return LivenessResult{}, err
	}
	return LivenessResult{
		Result:        result,
		Status:        vote.Status,
		MinScore:      vote.MinScore,
		LiveRatio:     vote.LiveRatio,
		FrameScores:   vote.FrameScores,
		LiveFrames:    vote.LiveFrames,
		AttackFrames:  vote.AttackFrames,
		ActionChecked: actionChecked,
		ActionPassed:  actionPassed,
		ActionCode:    actionCode,
		ActionDetail:  actionDetail,
	}, nil
}
