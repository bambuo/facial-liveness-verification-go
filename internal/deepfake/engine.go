// Package deepfake 提供 EfficientNet 深伪分类模型推理。
package deepfake

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
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

const inputSize = 224

var imagenetMean = [3]float32{0.485, 0.456, 0.406}
var imagenetStd = [3]float32{0.229, 0.224, 0.225}

// Engine 是已编译的深伪检测模型。
type Engine struct {
	m    onnx.Model
	exec *model.Exec
	mu   sync.Mutex
}

// Result 是深伪检测结果。
type Result struct {
	IsReal          bool    `json:"isReal"`
	RealProbability float64 `json:"realProbability"`
	FakeProbability float64 `json:"fakeProbability"`
}

// New 加载并编译深伪模型。
func New(path string) (*Engine, error) {
	m, err := parser.ParseFile(path)
	if err != nil {
		return nil, err
	}
	// FusedQuantizedDense requires a static contracting axis; the deepfake
	// model uses dynamic input dimensions, so keep the unfused conversion.
	m.DisableFusion()
	store := model.NewStore()
	if err := m.VariablesToScope(store.RootScope()); err != nil {
		_ = m.Close()
		return nil, err
	}
	inputs, inputShapes := m.Inputs()
	if len(inputs) == 0 {
		_ = m.Close()
		return nil, fmt.Errorf("深伪模型没有输入")
	}
	exec := model.MustNewExec(gobackend.GetBackend(), store, func(scope *model.Scope, input *graph.Node) []*graph.Node {
		return m.CallGraph(scope, input.Graph(), map[string]*graph.Node{inputs[0]: input})
	})
	dims := append([]int(nil), inputShapes[0].Dimensions...)
	for i, dim := range dims {
		if dim < 0 {
			dims[i] = 1
		}
	}
	if len(dims) == 4 {
		dims[0], dims[1], dims[2], dims[3] = 1, 3, inputSize, inputSize
	}
	if _, err := exec.Compile(shapes.Make(inputShapes[0].DType, dims...)); err != nil {
		exec.Finalize()
		_ = m.Close()
		return nil, err
	}
	return &Engine{m: m, exec: exec}, nil
}

// Close 释放模型资源。
func (e *Engine) Close() {
	if e == nil || e.exec == nil {
		return
	}
	e.exec.Finalize()
	_ = e.m.Close()
	e.exec = nil
}

// Detect 检测人脸裁剪是否真实。
func (e *Engine) Detect(img image.Image) (Result, error) {
	if e == nil || e.exec == nil {
		return Result{}, fmt.Errorf("深伪模型未启用")
	}
	rgb := resizeTo224(img)
	input := make([]float32, 3*inputSize*inputSize)
	pixels := inputSize * inputSize
	for y := 0; y < inputSize; y++ {
		for x := 0; x < inputSize; x++ {
			r, g, b, _ := rgb.At(x, y).RGBA()
			input[y*inputSize+x] = (float32(r)/65535 - imagenetMean[0]) / imagenetStd[0]
			input[pixels+y*inputSize+x] = (float32(g)/65535 - imagenetMean[1]) / imagenetStd[1]
			input[2*pixels+y*inputSize+x] = (float32(b)/65535 - imagenetMean[2]) / imagenetStd[2]
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	out, err := e.exec.Call(tensors.FromFlatDataAndDimensions(input, 1, 3, inputSize, inputSize))
	if err != nil {
		return Result{}, err
	}
	if len(out) == 0 {
		return Result{}, fmt.Errorf("深伪模型没有输出")
	}
	defer out[0].FinalizeAll()
	var logits []float32
	out[0].MustConstFlatData(func(value any) { logits = value.([]float32) })
	if len(logits) < 2 {
		return Result{}, fmt.Errorf("深伪输出维度异常")
	}
	maxLogit := max(logits[0], logits[1])
	fake := math.Exp(float64(logits[0] - maxLogit))
	real := math.Exp(float64(logits[1] - maxLogit))
	total := fake + real
	return Result{
		IsReal: real > fake, RealProbability: real / total, FakeProbability: fake / total,
	}, nil
}

// CropFace 按检测框外扩 20% 并裁剪，供深伪模型输入。
func CropFace(data []byte, box domain.FaceBox) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	marginX := (box.X2 - box.X1) * 0.2
	marginY := (box.Y2 - box.Y1) * 0.2
	left := max(0, int(box.X1-marginX))
	top := max(0, int(box.Y1-marginY))
	right := min(bounds.Dx(), int(box.X2+marginX))
	bottom := min(bounds.Dy(), int(box.Y2+marginY))
	if right <= left || bottom <= top {
		return nil, fmt.Errorf("人脸裁剪区域无效")
	}
	dst := image.NewRGBA(image.Rect(0, 0, right-left, bottom-top))
	draw.Draw(dst, dst.Bounds(), img, image.Pt(left, top), draw.Src)
	return dst, nil
}

func resizeTo224(img image.Image) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, inputSize, inputSize))
	bounds := img.Bounds()
	for y := 0; y < inputSize; y++ {
		for x := 0; x < inputSize; x++ {
			sx := bounds.Min.X + x*bounds.Dx()/inputSize
			sy := bounds.Min.Y + y*bounds.Dy()/inputSize
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	return dst
}
