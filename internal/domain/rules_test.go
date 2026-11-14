package domain

import (
	"math"
	"math/rand"
	"testing"
)

func TestBuildActions(t *testing.T) {
	p := Policy{ActionsCount: 3, ActionVerifyEnabled: true}
	actions := p.BuildActions(rand.New(rand.NewSource(1)))
	if len(actions) != 3 || actions[0] != "straight" {
		t.Fatalf("unexpected actions: %v", actions)
	}
	seen := map[string]bool{}
	for _, a := range actions {
		if seen[a] {
			t.Fatalf("duplicate action: %v", actions)
		}
		seen[a] = true
	}
}

func TestApplyPolicyOverrideAndClamp(t *testing.T) {
	base := Policy{FramesPerAction: 5, ActionsCount: 2, MatchThreshold: 0.35}
	got := ApplyPolicy(base, []byte(`{"match_threshold":0.5,"frames_per_action":20,"actions_count":0}`))
	if got.MatchThreshold != 0.5 {
		t.Fatalf("match threshold = %v", got.MatchThreshold)
	}
	if got.FramesPerAction != 8 || got.ActionsCount != 1 {
		t.Fatalf("clamp failed: %+v", got)
	}
}

func TestSatisfies(t *testing.T) {
	p := Policy{YawThreshold: 0.2, StraightYaw: 0.15, MouthOpenRatio: 1.08}
	if !Satisfies("straight", FrameMetric{FaceFound: true, YawRatio: 0.05}, nil, p) {
		t.Fatal("straight should pass")
	}
	if Satisfies("straight", FrameMetric{FaceFound: true, YawRatio: 0.4}, nil, p) {
		t.Fatal("straight should fail")
	}
	if !Satisfies("turn_left", FrameMetric{FaceFound: true, YawRatio: -0.3}, nil, p) {
		t.Fatal("turn_left should pass")
	}
	baseline := 0.4
	if !Satisfies("mouth_open", FrameMetric{FaceFound: true, MouthRatio: 0.5}, &baseline, p) {
		t.Fatal("mouth_open should pass")
	}
	if Satisfies("turn_right", FrameMetric{FaceFound: false, YawRatio: 1}, nil, p) {
		t.Fatal("no face should fail")
	}
}

func TestAggregateLiveness(t *testing.T) {
	p := Policy{LiveRatioThreshold: 0.5}
	if v, ok := AggregateLiveness(5, 0, 0.9, p); !ok || v.Status != LivenessReal {
		t.Fatalf("expected real: %+v", v)
	}
	if v, ok := AggregateLiveness(4, 1, 0.1, p); !ok || v.Status != LivenessSpoof {
		t.Fatalf("expected spoof: %+v", v)
	}
	if v, ok := AggregateLiveness(1, 1, 0.9, p); !ok || v.Status != LivenessSpoof {
		t.Fatalf("any attack frame should be spoof: %+v", v)
	}
	if _, ok := AggregateLiveness(0, 0, 1, p); ok {
		t.Fatal("judged=0 should be not ok")
	}
}

func TestCompareAndRound(t *testing.T) {
	score, err := Compare([]float32{1, 0}, []float32{1, 0})
	if err != nil || math.Abs(score-1) > 1e-9 {
		t.Fatalf("compare = %v, %v", score, err)
	}
	if _, err := Compare([]float32{1}, []float32{1, 2}); err == nil {
		t.Fatal("dimension mismatch should fail")
	}
	if got := RoundAway(0.123456, 5); got != 0.12346 {
		t.Fatalf("round = %v", got)
	}
}
