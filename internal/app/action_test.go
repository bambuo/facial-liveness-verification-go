package app

import (
	"context"
	"errors"
	"testing"

	"facial-liveness-verification-go/internal/domain"
)

type fakeEngine struct{ metrics []domain.FrameMetric }

func (f *fakeEngine) ModelVersion() string                       { return "fake-1.0" }
func (f *fakeEngine) DetectAll([]byte) ([]domain.FaceBox, error) { return nil, nil }
func (f *fakeEngine) ExtractLargest([][]byte) ([]float32, error) { return nil, nil }
func (f *fakeEngine) CheckLiveness([][]byte, float64, float64) (domain.LivenessVote, error) {
	return domain.LivenessVote{Status: domain.LivenessReal, MinScore: 1, LiveRatio: 1}, nil
}
func (f *fakeEngine) Analyze([][]byte) ([]domain.FrameMetric, error) { return f.metrics, nil }
func (f *fakeEngine) Compare(a, b []float32) (float64, error)        { return 0, nil }

func testPolicy() domain.Policy {
	return domain.Policy{YawThreshold: 0.2, StraightYaw: 0.15, MouthOpenRatio: 1.08, ActionPassRatio: 0.5}
}

func frames(n int) [][]byte {
	out := make([][]byte, n)
	for i := range out {
		out[i] = []byte{byte(i)}
	}
	return out
}

func TestCheckActionNoFrames(t *testing.T) {
	svc := NewActionService(&fakeEngine{})
	ok, detail, err := svc.CheckAction(context.Background(), "straight", nil, nil, testPolicy())
	if err != nil || ok || detail != "无帧" {
		t.Fatalf("got %v %q %v", ok, detail, err)
	}
}

func TestCheckActionNoFace(t *testing.T) {
	svc := NewActionService(&fakeEngine{metrics: []domain.FrameMetric{{FaceFound: false}}})
	ok, detail, err := svc.CheckAction(context.Background(), "straight", frames(1), nil, testPolicy())
	if err != nil || ok || detail != "NO_FACE" {
		t.Fatalf("got %v %q %v", ok, detail, err)
	}
}

func TestVerifyStraightPass(t *testing.T) {
	metrics := make([]domain.FrameMetric, 5)
	for i := range metrics {
		metrics[i] = domain.FrameMetric{FaceFound: true, YawRatio: 0.05, MouthRatio: 0.4}
	}
	svc := NewActionService(&fakeEngine{metrics: metrics})
	if err := svc.Verify(context.Background(), []string{"straight"}, 5, frames(5), testPolicy()); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyStraightFails(t *testing.T) {
	metrics := make([]domain.FrameMetric, 5)
	for i := range metrics {
		metrics[i] = domain.FrameMetric{FaceFound: true, YawRatio: 0.4, MouthRatio: 0.4}
	}
	svc := NewActionService(&fakeEngine{metrics: metrics})
	err := svc.Verify(context.Background(), []string{"straight"}, 5, frames(5), testPolicy())
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeActionMismatch {
		t.Fatalf("expected ACTION_MISMATCH, got %v", err)
	}
}

func TestVerifyMouthRelativeToBaseline(t *testing.T) {
	var metrics []domain.FrameMetric
	for i := 0; i < 5; i++ {
		metrics = append(metrics, domain.FrameMetric{FaceFound: true, YawRatio: 0.05, MouthRatio: 0.4})
	}
	for i := 0; i < 5; i++ {
		metrics = append(metrics, domain.FrameMetric{FaceFound: true, YawRatio: 0.05, MouthRatio: 0.6})
	}
	svc := NewActionService(&fakeEngine{metrics: metrics})
	if err := svc.Verify(context.Background(), []string{"straight", "mouth_open"}, 5, frames(10), testPolicy()); err != nil {
		t.Fatal(err)
	}
}

func TestVerifyFrameCountMismatch(t *testing.T) {
	svc := NewActionService(&fakeEngine{})
	err := svc.Verify(context.Background(), []string{"straight", "turn_left"}, 5, frames(7), testPolicy())
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeAuthInvalid {
		t.Fatalf("expected AUTH_INVALID, got %v", err)
	}
}

func TestVerifyNoFaceInPhase(t *testing.T) {
	metrics := make([]domain.FrameMetric, 5)
	svc := NewActionService(&fakeEngine{metrics: metrics})
	err := svc.Verify(context.Background(), []string{"straight"}, 5, frames(5), testPolicy())
	var de *domain.Error
	if !errors.As(err, &de) || de.Code != domain.CodeNoFace {
		t.Fatalf("expected NO_FACE, got %v", err)
	}
}
