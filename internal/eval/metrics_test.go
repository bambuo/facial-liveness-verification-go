package eval

import (
	"math"
	"testing"
)

func TestSweepAndEER(t *testing.T) {
	genuine := []float64{0.9, 0.8}
	impostor := []float64{0.2, 0.3}
	points := Sweep(genuine, impostor, []float64{0.5, 0.85, 1.0001})
	if points[0].FAR != 0 || points[0].FRR != 0 {
		t.Fatalf("t=0.5: %+v", points[0])
	}
	if points[1].FAR != 0 || points[1].FRR != 0.5 {
		t.Fatalf("t=0.85: %+v", points[1])
	}
	if points[2].FAR != 0 || points[2].FRR != 1 {
		t.Fatalf("t=1.0001: %+v", points[2])
	}
	eer, threshold := EER(points)
	if eer != 0 || threshold != 0.5 {
		t.Fatalf("eer=%v threshold=%v", eer, threshold)
	}
}

func TestROCAUC(t *testing.T) {
	auc := ROCAUC([]float64{0.9, 0.8}, []float64{0.2, 0.3})
	if math.Abs(auc-1) > 1e-9 {
		t.Fatalf("auc=%v", auc)
	}
	auc = ROCAUC([]float64{0.5}, []float64{0.5})
	if auc != 0.5 {
		t.Fatalf("auc=%v", auc)
	}
}

func TestRecommendThreshold(t *testing.T) {
	genuine := []float64{0.9, 0.8}
	impostor := []float64{0.2, 0.3}
	thresholds := DefaultThresholds(0.01)
	threshold, frr, ok := RecommendThreshold(genuine, impostor, thresholds, 0)
	if !ok || frr != 0 || math.Abs(threshold-0.31) > 1e-9 {
		t.Fatalf("threshold=%v frr=%v ok=%v", threshold, frr, ok)
	}
}

func TestLivenessSweep(t *testing.T) {
	bona := []float64{0.9, 0.8}
	attack := []float64{0.1, 0.85}
	points := LivenessSweep(bona, attack, []float64{0.8, 0.9})
	if points[0].APCER != 0.5 || points[0].BPCER != 0 {
		t.Fatalf("t=0.8: %+v", points[0])
	}
	if points[1].APCER != 0 || points[1].BPCER != 0.5 {
		t.Fatalf("t=0.9: %+v", points[1])
	}
	min, mean := MinMean([]float64{0.8, 0.9, 1.0})
	if min != 0.8 || math.Abs(mean-0.9) > 1e-9 {
		t.Fatalf("min=%v mean=%v", min, mean)
	}
}
