package eval

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// WriteOutputs 将报告与明细写入目录。
func WriteOutputs(dir string, report *Report, pairs []Pair, samples []LivenessSample) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "report.json"), report); err != nil {
		return err
	}
	if err := writePairs(filepath.Join(dir, "scores.csv"), pairs); err != nil {
		return err
	}
	if err := writeThresholds(filepath.Join(dir, "roc.csv"), report.Thresholds); err != nil {
		return err
	}
	if err := writeLiveness(filepath.Join(dir, "liveness.csv"), samples); err != nil {
		return err
	}
	if report.Liveness != nil {
		if err := writeLivenessMetrics(filepath.Join(dir, "liveness_metrics.csv"), report.Liveness.Thresholds); err != nil {
			return err
		}
	}
	return nil
}

func writeJSON(path string, v any) error {
	raw, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func writePairs(path string, pairs []Pair) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"kind", "id1", "image1", "id2", "image2", "score"}); err != nil {
		return err
	}
	for _, p := range pairs {
		if err := w.Write([]string{p.Kind, p.ID1, p.Image1, p.ID2, p.Image2, formatFloat(p.Score)}); err != nil {
			return err
		}
	}
	return w.Error()
}

func writeThresholds(path string, points []ThresholdPoint) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"threshold", "far", "frr", "tar"}); err != nil {
		return err
	}
	for _, p := range points {
		if err := w.Write([]string{formatFloat(p.Threshold), formatFloat(p.FAR), formatFloat(p.FRR), formatFloat(p.TAR)}); err != nil {
			return err
		}
	}
	return w.Error()
}

func writeLiveness(path string, samples []LivenessSample) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"id", "label", "frames", "min_score", "mean_score", "error"}); err != nil {
		return err
	}
	for _, s := range samples {
		if err := w.Write([]string{s.ID, s.Label, strconv.Itoa(s.Frames), formatFloat(s.MinScore), formatFloat(s.MeanScore), s.Error}); err != nil {
			return err
		}
	}
	return w.Error()
}

func writeLivenessMetrics(path string, points []LivenessPoint) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	if err := w.Write([]string{"threshold", "apcer", "bpcer"}); err != nil {
		return err
	}
	for _, p := range points {
		if err := w.Write([]string{formatFloat(p.Threshold), formatFloat(p.APCER), formatFloat(p.BPCER)}); err != nil {
			return err
		}
	}
	return w.Error()
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', 6, 64)
}
