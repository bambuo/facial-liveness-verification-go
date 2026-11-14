package app

import (
	"context"
	"fmt"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/port"
)

// ActionService 执行单动作预检与整单动作校验。
type ActionService struct {
	engine port.FaceEngine
}

// NewActionService 构造动作校验用例。
func NewActionService(engine port.FaceEngine) *ActionService {
	return &ActionService{engine: engine}
}

// CheckAction 判断单动作近期帧是否达标，不消费 challenge、不写库。
func (s *ActionService) CheckAction(ctx context.Context, action string, frames, baselineFrames [][]byte, policy domain.Policy) (bool, string, error) {
	_ = ctx
	if len(frames) == 0 {
		return false, "无帧", nil
	}
	metrics, err := s.engine.Analyze(frames)
	if err != nil {
		return false, "", err
	}
	withFace := 0
	for _, m := range metrics {
		if m.FaceFound {
			withFace++
		}
	}
	if withFace == 0 {
		return false, "NO_FACE", nil
	}
	baseline := s.baseline(baselineFrames)
	passed := 0
	for _, m := range metrics {
		if domain.Satisfies(action, m, baseline, policy) {
			passed++
		}
	}
	satisfied := passed >= domain.RequiredPasses(len(metrics), policy.ActionPassRatio)
	return satisfied, domain.ActionStats(metrics), nil
}

// Verify 整单校验：帧数必须与动作序列精确匹配，逐段判定。
func (s *ActionService) Verify(ctx context.Context, actions []string, framesPerAction int, frames [][]byte, policy domain.Policy) error {
	_ = ctx
	if len(actions) == 0 || framesPerAction <= 0 {
		return nil
	}
	if len(frames) != len(actions)*framesPerAction {
		return domain.New(domain.CodeAuthInvalid, "帧数与动作序列不匹配")
	}
	metrics, err := s.engine.Analyze(frames)
	if err != nil {
		return err
	}
	var mouthBaseline *float64
	for phase, action := range actions {
		from := phase * framesPerAction
		segment := metrics[from : from+framesPerAction]
		withFace := 0
		for _, m := range segment {
			if m.FaceFound {
				withFace++
			}
		}
		if withFace == 0 {
			return domain.Of(domain.CodeNoFace)
		}
		if action == "straight" {
			mouthBaseline = minMouth(segment)
		}
		passed := 0
		for _, m := range segment {
			if domain.Satisfies(action, m, mouthBaseline, policy) {
				passed++
			}
		}
		if passed < domain.RequiredPasses(framesPerAction, policy.ActionPassRatio) {
			return domain.New(domain.CodeActionMismatch, fmt.Sprintf("动作未完成：%s（%s）", action, domain.ActionStats(segment)))
		}
	}
	return nil
}

func (s *ActionService) baseline(frames [][]byte) *float64 {
	if len(frames) == 0 {
		return nil
	}
	metrics, err := s.engine.Analyze(frames)
	if err != nil {
		return nil
	}
	return minMouth(metrics)
}

func minMouth(metrics []domain.FrameMetric) *float64 {
	found := false
	minV := 0.0
	for _, m := range metrics {
		if !m.FaceFound {
			continue
		}
		if !found || m.MouthRatio < minV {
			minV = m.MouthRatio
			found = true
		}
	}
	if !found {
		zero := 0.0
		return &zero
	}
	return &minV
}
