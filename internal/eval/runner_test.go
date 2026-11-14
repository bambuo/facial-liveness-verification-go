package eval

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"facial-liveness-verification-go/internal/domain"
)

type fakeEngine struct{}

func (fakeEngine) ModelVersion() string { return "fake-1.0" }

func (fakeEngine) ExtractLargest(frames [][]byte) ([]float32, error) {
	return embedding(frames[0]), nil
}

func (fakeEngine) Compare(left, right []float32) (float64, error) {
	return float64(left[0]*right[0] + left[1]*right[1]), nil
}

func (fakeEngine) CheckLiveness(frames [][]byte, _, _ float64) (domain.LivenessVote, error) {
	score := 0.9
	if frames[0][0] == 'b' {
		score = 0.1
	}
	return domain.LivenessVote{Status: domain.LivenessReal, MinScore: score, LiveRatio: 1}, nil
}

func embedding(data []byte) []float32 {
	if len(data) > 0 && data[0] == 'b' {
		return []float32{0, 1}
	}
	return []float32{1, 0}
}

func TestRunnerMetricsAndOutputs(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.yaml")
	manifest := `
identities:
  - id: A
    images: [a1.jpg, a2.jpg]
  - id: B
    images: [b1.jpg, b2.jpg]
liveness:
  bona_fide:
    - id: liveA
      images: [a1.jpg, a2.jpg]
  attacks:
    - id: attackB
      label: print
      images: [b1.jpg]
options:
  threshold_step: 0.01
`
	if err := os.WriteFile(manifestPath, []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	runner := NewRunner(fakeEngine{}, m)
	runner.ImageLoader = func(path string) ([]byte, error) {
		base := filepath.Base(path)
		return []byte{base[0]}, nil
	}
	runner.TargetFAR = 0
	report, pairs, samples, err := runner.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if report.Counts.Identities != 2 || report.Counts.Embedded != 4 {
		t.Fatalf("counts: %+v", report.Counts)
	}
	if report.Counts.GenuinePairs != 2 || report.Counts.ImpostorPairs != 4 {
		t.Fatalf("pairs: %+v", report.Counts)
	}
	if len(pairs) != 6 || len(samples) != 2 {
		t.Fatalf("pairs=%d samples=%d", len(pairs), len(samples))
	}
	if math.Abs(report.EER) > 1e-9 || math.Abs(report.AUC-1) > 1e-9 {
		t.Fatalf("eer=%v auc=%v", report.EER, report.AUC)
	}
	if !report.TargetAchieved || report.FRRAtTarget != 0 {
		t.Fatalf("threshold: %+v", report)
	}
	if report.Liveness == nil || report.Liveness.ServicePoint.APCER != 0 || report.Liveness.ServicePoint.BPCER != 0 {
		t.Fatalf("liveness: %+v", report.Liveness)
	}

	out := filepath.Join(dir, "out")
	if err := WriteOutputs(out, report, pairs, samples); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"report.json", "scores.csv", "roc.csv", "liveness.csv", "liveness_metrics.csv"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}
