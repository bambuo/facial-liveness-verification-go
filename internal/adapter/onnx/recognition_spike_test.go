package onnx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"os"
	"sort"
	"testing"
	"time"

	"facial-liveness-verification-go/internal/domain"
	"facial-liveness-verification-go/internal/eval"
)

// recSpikeEngine 在同一对齐与识别链路上替换检测模型，用于 EER 与耗时对比。
type recSpikeEngine struct {
	det    *Engine
	scrfd  *scrfdDetector
	rec    *execModel
	aura   bool
	dim    int
	detMS  []float64
	recMS  []float64
	total  []float64
}

// ModelVersion 返回本次评测的识别模型标识。
func (e *recSpikeEngine) ModelVersion() string {
	if e.aura {
		return "auraface-glintr100-512d"
	}
	return modelVersion
}

// ExtractLargest 检测最大人脸、对齐后提取特征。
func (e *recSpikeEngine) ExtractLargest(frames [][]byte) ([]float32, error) {
	var best []float32
	bestArea := -1.0
	for _, frame := range frames {
		totalStart := time.Now()
		img, err := decodeImage(frame)
		if err != nil {
			return nil, err
		}
		boxes, err := e.detectFaces(img)
		if err != nil {
			return nil, err
		}
		box := largestFace(boxes)
		if box == nil {
			continue
		}
		area := (box.X2 - box.X1) * (box.Y2 - box.Y1)
		if area <= bestArea {
			continue
		}
		if !box.HasLandmarks {
			return nil, domain.New(domain.CodeNoFace, "关键点缺失，无法对齐")
		}
		aligned := alignCrop(img, box.Landmarks)
		input := toCHW(aligned, false)
		if e.aura {
			for i := range input {
				input[i] = (input[i] - 127.5) / 127.5
			}
		}
		recStart := time.Now()
		out, err := e.rec.run(input)
		e.recMS = append(e.recMS, float64(time.Since(recStart).Microseconds())/1000)
		if err != nil {
			return nil, err
		}
		feat := flatF32(out[0])
		l2Normalize(feat)
		best, bestArea = feat, area
		e.dim = len(feat)
		e.total = append(e.total, float64(time.Since(totalStart).Microseconds())/1000)
	}
	if best == nil {
		return nil, domain.Of(domain.CodeNoFace)
	}
	return best, nil
}

// detectFaces 按当前检测模式执行检测并记录耗时。
func (e *recSpikeEngine) detectFaces(img *image.RGBA) ([]domain.FaceBox, error) {
	start := time.Now()
	var boxes []domain.FaceBox
	var err error
	if e.scrfd != nil {
		boxes, err = e.scrfd.Detect(img, 0.5, 0.4)
	} else {
		boxes, err = e.det.detect(img, e.det.cfg.AnalysisThreshold)
	}
	e.detMS = append(e.detMS, float64(time.Since(start).Microseconds())/1000)
	return boxes, err
}

// Compare 计算两个特征的余弦相似度。
func (e *recSpikeEngine) Compare(left, right []float32) (float64, error) {
	return domain.Compare(left, right)
}

// CheckLiveness 本次只评测 1:1，不支持活体。
func (e *recSpikeEngine) CheckLiveness([][]byte, float64, float64) (domain.LivenessVote, error) {
	return domain.LivenessVote{}, errors.New("recognition spike 不支持活体")
}

type latencyReport struct {
	Mode           string  `json:"mode"`
	Detector       string  `json:"detector"`
	Engine         string  `json:"engine"`
	EmbeddingDim   int     `json:"embedding_dim"`
	Images         int     `json:"images"`
	LoadMS         float64 `json:"load_ms"`
	DetectP50MS    float64 `json:"detect_p50_ms"`
	DetectP95MS    float64 `json:"detect_p95_ms"`
	RecP50MS       float64 `json:"rec_p50_ms"`
	RecP95MS       float64 `json:"rec_p95_ms"`
	RecMeanMS      float64 `json:"rec_mean_ms"`
	TotalP50MS     float64 `json:"total_p50_ms"`
	TotalP95MS     float64 `json:"total_p95_ms"`
	TotalMeanMS    float64 `json:"total_mean_ms"`
	EER            float64 `json:"eer"`
	AUC            float64 `json:"auc"`
	TargetFAR      float64 `json:"target_far"`
	Threshold      float64 `json:"threshold"`
	FRRAtTarget    float64 `json:"frr_at_target"`
	TargetAchieved bool    `json:"target_achieved"`
}

// percentile 返回升序样本的最近秩百分位。
func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

// mean 返回样本均值。
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// TestRecognitionSpike 对比可替换识别模型的 EER 与单图耗时；默认跳过。
func TestRecognitionSpike(t *testing.T) {
	if os.Getenv("REC_SPIKE") != "1" {
		t.Skip("设置 REC_SPIKE=1 运行识别模型对比")
	}
	modelDir := os.Getenv("MODEL_DIR")
	manifestPath := os.Getenv("LFW_MANIFEST")
	outDir := os.Getenv("OUT_DIR")
	mode := os.Getenv("REC_MODE")
	detMode := os.Getenv("DET_MODE")
	if detMode == "" {
		detMode = "yunet"
	}
	if modelDir == "" || manifestPath == "" || outDir == "" || mode == "" {
		t.Fatal("需要 MODEL_DIR、LFW_MANIFEST、OUT_DIR、REC_MODE")
	}

	loadStart := time.Now()
	var err error
	var det *execModel
	var scrfd *scrfdDetector
	switch detMode {
	case "yunet":
		det, err = loadModel(modelDir+"/face_detection_yunet_2023mar.onnx", yunetInputSize)
		if err != nil {
			t.Fatalf("加载检测模型失败: %v", err)
		}
		defer det.close()
	case "scrfd":
		path := os.Getenv("SCRFD_MODEL")
		if path == "" {
			t.Fatal("scrfd 模式需要 SCRFD_MODEL")
		}
		scrfd, err = newSCRFD(path)
		if err != nil {
			t.Fatalf("加载 SCRFD 失败: %v", err)
		}
		defer scrfd.close()
	default:
		t.Fatalf("未知 DET_MODE: %s", detMode)
	}

	var recPath string
	switch mode {
	case "aura":
		recPath = os.Getenv("AURAFACE_MODEL")
		if recPath == "" {
			t.Fatal("aura 模式需要 AURAFACE_MODEL")
		}
	case "sface":
		recPath = modelDir + "/face_recognition_sface_2021dec.onnx"
	default:
		t.Fatalf("未知 REC_MODE: %s", mode)
	}
	rec, err := loadModel(recPath, 112)
	if err != nil {
		t.Fatalf("加载识别模型失败: %v", err)
	}
	defer rec.close()
	loadMS := float64(time.Since(loadStart).Microseconds()) / 1000

	manifest, err := eval.LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("加载清单失败: %v", err)
	}
	engine := &recSpikeEngine{
		det: &Engine{
			cfg: Config{DetectionThreshold: 0.9, NMSThreshold: 0.3, AnalysisThreshold: 0.6},
			det: det,
		},
		scrfd: scrfd,
		rec:   rec,
		aura:  mode == "aura",
	}
	runner := eval.NewRunner(engine, manifest)
	runner.TargetFAR = 0.001
	report, pairs, samples, err := runner.Run(context.Background())
	if err != nil {
		t.Fatalf("评测失败: %v", err)
	}
	if err := eval.WriteOutputs(outDir, report, pairs, samples); err != nil {
		t.Fatalf("写报告失败: %v", err)
	}

	lat := latencyReport{
		Mode:           mode,
		Detector:       detMode,
		Engine:         engine.ModelVersion(),
		EmbeddingDim:   engine.dim,
		Images:         report.Counts.Embedded,
		LoadMS:         loadMS,
		DetectP50MS:    percentile(engine.detMS, 0.50),
		DetectP95MS:    percentile(engine.detMS, 0.95),
		RecP50MS:       percentile(engine.recMS, 0.50),
		RecP95MS:       percentile(engine.recMS, 0.95),
		RecMeanMS:      mean(engine.recMS),
		TotalP50MS:     percentile(engine.total, 0.50),
		TotalP95MS:     percentile(engine.total, 0.95),
		TotalMeanMS:    mean(engine.total),
		EER:            report.EER,
		AUC:            report.AUC,
		TargetFAR:      report.TargetFAR,
		Threshold:      report.RecommendedThreshold,
		FRRAtTarget:    report.FRRAtTarget,
		TargetAchieved: report.TargetAchieved,
	}
	raw, err := json.MarshalIndent(lat, "", "  ")
	if err != nil {
		t.Fatalf("编码耗时报告失败: %v", err)
	}
	if err := os.WriteFile(outDir+"/latency.json", raw, 0o644); err != nil {
		t.Fatalf("写耗时报告失败: %v", err)
	}

	fmt.Printf("mode=%s det=%s engine=%s dim=%d images=%d load_ms=%.1f\n", lat.Mode, lat.Detector, lat.Engine, lat.EmbeddingDim, lat.Images, lat.LoadMS)
	fmt.Printf("detect p50=%.1fms p95=%.1fms\n", lat.DetectP50MS, lat.DetectP95MS)
	fmt.Printf("rec p50=%.1fms p95=%.1fms mean=%.1fms\n", lat.RecP50MS, lat.RecP95MS, lat.RecMeanMS)
	fmt.Printf("total p50=%.1fms p95=%.1fms mean=%.1fms\n", lat.TotalP50MS, lat.TotalP95MS, lat.TotalMeanMS)
	fmt.Printf("EER=%.6f threshold=%.4f AUC=%.6f\n", lat.EER, lat.Threshold, lat.AUC)
	fmt.Printf("target_far=%.6f threshold=%.4f FRR=%.6f achieved=%v\n",
		lat.TargetFAR, lat.Threshold, lat.FRRAtTarget, lat.TargetAchieved)
	fmt.Printf("report=%s/report.json\n", outDir)
}
