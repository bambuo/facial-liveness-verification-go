package onnx

import (
	"fmt"
	"image"
	"math"
	"path/filepath"
	"sync"

	"github.com/gomlx/compute/gobackend"
	"github.com/gomlx/compute/shapes"
	"github.com/gomlx/gomlx/core/graph"
	"github.com/gomlx/gomlx/core/tensors"
	"github.com/gomlx/gomlx/ml/model"
	"github.com/gomlx/onnx-gomlx/onnx"
	"github.com/gomlx/onnx-gomlx/onnx/parser"

	"facial-liveness-verification-go/internal/domain"
)

const (
	modelVersion   = "mica-sface-128d-1.0.2"
	yunetInputSize = 640
	fasInputSize   = 80
	liveIndex      = 1
)

// Config 是引擎启动参数。
type Config struct {
	ModelDir           string
	DetectionThreshold float64
	NMSThreshold       float64
	AnalysisThreshold  float64
	LivenessCropScale  float64
}

// Engine 用 onnx-gomlx + GoMLX 纯 Go 后端执行三个模型。
type Engine struct {
	cfg Config
	det *execModel
	rec *execModel
	fas *execModel
}

// New 加载 YuNet、SFace、MiniFASNetV2 并编译固定 batch=1 的计算图。
func New(cfg Config) (*Engine, error) {
	det, err := loadModel(filepath.Join(cfg.ModelDir, "face_detection_yunet_2023mar.onnx"), yunetInputSize)
	if err != nil {
		return nil, fmt.Errorf("load yunet: %w", err)
	}
	rec, err := loadModel(filepath.Join(cfg.ModelDir, "face_recognition_sface_2021dec.onnx"), 112)
	if err != nil {
		det.close()
		return nil, fmt.Errorf("load sface: %w", err)
	}
	fas, err := loadModel(filepath.Join(cfg.ModelDir, "2.7_80x80_MiniFASNetV2.onnx"), fasInputSize)
	if err != nil {
		det.close()
		rec.close()
		return nil, fmt.Errorf("load minifas: %w", err)
	}
	return &Engine{cfg: cfg, det: det, rec: rec, fas: fas}, nil
}

// Close 释放模型资源。
func (e *Engine) Close() {
	e.det.close()
	e.rec.close()
	e.fas.close()
}

// ModelVersion 返回引擎特征版本。
func (e *Engine) ModelVersion() string { return modelVersion }

// DetectAll 检测图中全部人脸。
func (e *Engine) DetectAll(data []byte) ([]domain.FaceBox, error) {
	img, err := decodeImage(data)
	if err != nil {
		return nil, err
	}
	return e.detect(img, e.cfg.DetectionThreshold)
}

// ExtractLargest 从多帧中取最大脸并提取 128 维特征。
func (e *Engine) ExtractLargest(frames [][]byte) ([]float32, error) {
	var best []float32
	bestArea := -1.0
	for _, frame := range frames {
		img, err := decodeImage(frame)
		if err != nil {
			return nil, err
		}
		boxes, err := e.detect(img, e.cfg.AnalysisThreshold)
		if err != nil {
			return nil, err
		}
		box := largestFace(boxes)
		if box == nil {
			continue
		}
		area := math.Max(0, box.X2-box.X1) * math.Max(0, box.Y2-box.Y1)
		if area <= bestArea {
			continue
		}
		if !box.HasLandmarks {
			return nil, domain.New(domain.CodeNoFace, "关键点缺失，无法对齐")
		}
		aligned := alignCrop(img, box.Landmarks)
		out, err := e.rec.run(toCHW(aligned, false))
		if err != nil {
			return nil, err
		}
		feat := flatF32(out[0])
		l2Normalize(feat)
		best, bestArea = feat, area
	}
	if best == nil {
		return nil, domain.Of(domain.CodeNoFace)
	}
	return best, nil
}

// CheckLiveness 多帧活体投票：任一攻击帧为 spoof，否则按 live 占比判 real/fuzzy。
func (e *Engine) CheckLiveness(frames [][]byte, threshold, liveRatioThreshold float64) (domain.LivenessVote, error) {
	live, attack := 0, 0
	minScore := 1.0
	for _, frame := range frames {
		img, err := decodeImage(frame)
		if err != nil {
			return domain.LivenessVote{}, err
		}
		boxes, err := e.detect(img, e.cfg.AnalysisThreshold)
		if err != nil {
			return domain.LivenessVote{}, err
		}
		box := largestFace(boxes)
		if box == nil {
			continue
		}
		crop, err := cropWithMargin(img, *box, e.cfg.LivenessCropScale, fasInputSize)
		if err != nil {
			return domain.LivenessVote{}, err
		}
		out, err := e.fas.run(toCHW(crop, true))
		if err != nil {
			return domain.LivenessVote{}, err
		}
		probs := softmax(flatF32(out[0]))
		score := probs[liveIndex]
		minScore = math.Min(minScore, score)
		if score >= threshold || argmax(probs) == liveIndex {
			live++
		} else {
			attack++
		}
	}
	vote, ok := domain.AggregateLiveness(live, attack, minScore, domain.Policy{LiveRatioThreshold: liveRatioThreshold})
	if !ok {
		return domain.LivenessVote{}, domain.Of(domain.CodeNoFace)
	}
	return vote, nil
}

// Analyze 逐帧计算 yawRatio 与 mouthRatio。
func (e *Engine) Analyze(frames [][]byte) ([]domain.FrameMetric, error) {
	out := make([]domain.FrameMetric, 0, len(frames))
	for _, frame := range frames {
		img, err := decodeImage(frame)
		if err != nil {
			return nil, err
		}
		boxes, err := e.detect(img, e.cfg.AnalysisThreshold)
		if err != nil {
			return nil, err
		}
		box := largestFace(boxes)
		if box == nil {
			out = append(out, domain.FrameMetric{})
			continue
		}
		out = append(out, metricsOf(*box))
	}
	return out, nil
}

// Compare 计算两个特征的余弦相似度。
func (e *Engine) Compare(left, right []float32) (float64, error) {
	return domain.Compare(left, right)
}

// execModel 是单个已编译模型。
type execModel struct {
	model    onnx.Model
	exec     *model.Exec
	input    string
	outputs  []string
	outIndex map[string]int
	shape    []int
	mu       sync.Mutex
}

func loadModel(path string, fixedSize int) (*execModel, error) {
	m, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}
	store := model.NewStore()
	if err := m.VariablesToScope(store.RootScope()); err != nil {
		m.Close()
		return nil, err
	}
	inNames, inShapes := m.Inputs()
	outNames, _ := m.Outputs()
	if len(inNames) != 1 {
		m.Close()
		return nil, fmt.Errorf("模型输入数量不是 1: %v", inNames)
	}
	dims := append([]int(nil), inShapes[0].Dimensions...)
	for i, d := range dims {
		if d < 0 {
			dims[i] = 1
		}
	}
	if len(dims) == 4 && dims[3] != fixedSize {
		dims[3] = fixedSize
		dims[2] = fixedSize
	}
	inputName := inNames[0]
	exec := model.MustNewExec(gobackend.GetBackend(), store, func(scope *model.Scope, input *graph.Node) []*graph.Node {
		return m.CallGraph(scope, input.Graph(), map[string]*graph.Node{inputName: input})
	})
	if _, err := exec.Compile(shapes.Make(inShapes[0].DType, dims...)); err != nil {
		exec.Finalize()
		m.Close()
		return nil, err
	}
	idx := make(map[string]int, len(outNames))
	for i, name := range outNames {
		idx[name] = i
	}
	return &execModel{model: m, exec: exec, input: inputName, outputs: outNames, outIndex: idx, shape: dims}, nil
}

func (m *execModel) run(input []float32) ([]*tensors.Tensor, error) {
	t := tensors.FromFlatDataAndDimensions(input, m.shape...)
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.exec.Call(t)
}

func (m *execModel) get(outs []*tensors.Tensor, name string) []float32 {
	i, ok := m.outIndex[name]
	if !ok {
		return nil
	}
	return flatF32(outs[i])
}

func (m *execModel) close() {
	if m.exec != nil {
		m.exec.Finalize()
	}
	if m.model != nil {
		_ = m.model.Close()
	}
}

func flatF32(t *tensors.Tensor) []float32 {
	var out []float32
	t.MustConstFlatData(func(v any) { out = v.([]float32) })
	return out
}

// detect 执行 YuNet letterbox 推理、解码与 NMS。
func (e *Engine) detect(img *image.RGBA, threshold float64) ([]domain.FaceBox, error) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	scale := math.Min(float64(yunetInputSize)/float64(w), float64(yunetInputSize)/float64(h))
	invScale := 1 / scale
	canvas := letterbox(img, yunetInputSize)
	outs, err := e.det.run(toCHW(canvas, true))
	if err != nil {
		return nil, err
	}
	boxes := make([]domain.FaceBox, 0, 64)
	for _, stride := range []int{8, 16, 32} {
		grid := yunetInputSize / stride
		cls := e.det.get(outs, fmt.Sprintf("cls_%d", stride))
		obj := e.det.get(outs, fmt.Sprintf("obj_%d", stride))
		bbo := e.det.get(outs, fmt.Sprintf("bbox_%d", stride))
		kps := e.det.get(outs, fmt.Sprintf("kps_%d", stride))
		for k := 0; k < grid*grid; k++ {
			col, row := k%grid, k/grid
			score := javaClip(float64(cls[k]))
			iou := javaClip(float64(obj[k]))
			if s2 := math.Sqrt(score * iou); s2 < scale {
				score = s2
			}
			if score < threshold {
				continue
			}
			ax := (float64(col) + float64(bbo[k*4])) * float64(stride)
			ay := (float64(row) + float64(bbo[k*4+1])) * float64(stride)
			d1 := math.Exp(float64(bbo[k*4+2])) * float64(stride)
			d2 := math.Exp(float64(bbo[k*4+3])) * float64(stride)
			box := domain.FaceBox{
				X1:    javaClipMax((ax - d1/2) * invScale),
				Y1:    javaClipMax((ay - d2/2) * invScale),
				X2:    javaClipMax((ax + d1/2) * invScale),
				Y2:    javaClipMax((ay + d2/2) * invScale),
				Score: score,
			}
			finite := true
			for j := 0; j < 5; j++ {
				lx := ((float64(col) + float64(kps[k*10+j*2])) * float64(stride)) * invScale
				ly := ((float64(row) + float64(kps[k*10+j*2+1])) * float64(stride)) * invScale
				box.Landmarks[j] = [2]float64{lx, ly}
				if math.IsNaN(lx) || math.IsNaN(ly) || math.IsInf(lx, 0) || math.IsInf(ly, 0) {
					finite = false
				}
			}
			box.HasLandmarks = finite
			boxes = append(boxes, box)
		}
	}
	return nms(boxes, e.cfg.NMSThreshold), nil
}

// javaClip 复刻 Java 版 clip(v,a,b)=max(v,min(a,b))，此处 a=0,b=1。
func javaClip(v float64) float64 { return math.Max(v, math.Min(0, 1)) }

// javaClipMax 复刻基线坐标收紧：只保证下界为 0。
func javaClipMax(v float64) float64 { return math.Max(v, 0) }

func nms(boxes []domain.FaceBox, threshold float64) []domain.FaceBox {
	kept := make([]domain.FaceBox, 0, len(boxes))
	for _, box := range boxes {
		suppressed := false
		for _, k := range kept {
			if iou(box, k) > threshold {
				suppressed = true
				break
			}
		}
		if !suppressed {
			kept = append(kept, box)
		}
	}
	return kept
}

func iou(a, b domain.FaceBox) float64 {
	ax1, ay1, ax2, ay2 := math.Floor(a.X1), math.Floor(a.Y1), math.Floor(a.X2), math.Floor(a.Y2)
	bx1, by1, bx2, by2 := math.Floor(b.X1), math.Floor(b.Y1), math.Floor(b.X2), math.Floor(b.Y2)
	ix1, iy1 := math.Max(ax1, bx1), math.Max(ay1, by1)
	ix2, iy2 := math.Min(ax2, bx2), math.Min(ay2, by2)
	iw, ih := ix2-ix1, iy2-iy1
	if iw <= 0 || ih <= 0 {
		return 0
	}
	inter := iw * ih
	return inter / ((ax2-ax1)*(ay2-ay1) + (bx2-bx1)*(by2-by1) - inter)
}

func metricsOf(box domain.FaceBox) domain.FrameMetric {
	if !box.HasLandmarks {
		return domain.FrameMetric{FaceFound: true}
	}
	lm := box.Landmarks
	eyeMidX := (lm[0][0] + lm[1][0]) / 2
	eyeDist := math.Hypot(lm[0][0]-lm[1][0], lm[0][1]-lm[1][1])
	if eyeDist < 1e-3 {
		return domain.FrameMetric{FaceFound: true}
	}
	noseX, noseY := lm[2][0], lm[2][1]
	mouthMidX := (lm[3][0] + lm[4][0]) / 2
	mouthMidY := (lm[3][1] + lm[4][1]) / 2
	return domain.FrameMetric{
		FaceFound:  true,
		YawRatio:   (noseX - eyeMidX) / eyeDist,
		MouthRatio: math.Hypot(noseX-mouthMidX, noseY-mouthMidY) / eyeDist,
	}
}

func softmax(logits []float32) []float64 {
	maxLogit := math.Inf(-1)
	for _, v := range logits {
		maxLogit = math.Max(maxLogit, float64(v))
	}
	out := make([]float64, len(logits))
	sum := 0.0
	for i, v := range logits {
		out[i] = math.Exp(float64(v) - maxLogit)
		sum += out[i]
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

func argmax(values []float64) int {
	idx := 0
	maxV := math.Inf(-1)
	for i, v := range values {
		if v > maxV {
			maxV, idx = v, i
		}
	}
	return idx
}

func l2Normalize(vector []float32) {
	norm := 0.0
	for _, v := range vector {
		norm += float64(v) * float64(v)
	}
	norm = math.Sqrt(norm)
	if norm <= 0 {
		return
	}
	for i := range vector {
		vector[i] = float32(float64(vector[i]) / norm)
	}
}
