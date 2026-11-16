package onnx

import (
	"fmt"
	"image"
	"math"

	"github.com/gomlx/gomlx/core/tensors"

	"facial-liveness-verification-go/internal/domain"
)

// scrfdDetector 用 SCRFD-10G 输出 2 anchor/位置的分数、框距离与 5 点关键点。
type scrfdDetector struct {
	model *execModel
	size  int
}

// newSCRFD 加载固定 640×640 输入的 SCRFD 模型。
func newSCRFD(path string) (*scrfdDetector, error) {
	m, err := loadModel(path, 640)
	if err != nil {
		return nil, err
	}
	return &scrfdDetector{model: m, size: 640}, nil
}

// close 释放 SCRFD 模型。
func (d *scrfdDetector) close() {
	d.model.close()
}

// Detect 执行 SCRFD 推理并解码置信度过滤后的框与关键点。
func (d *scrfdDetector) Detect(img *image.RGBA, threshold, nmsThreshold float64) ([]domain.FaceBox, error) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	scale := math.Min(float64(d.size)/float64(w), float64(d.size)/float64(h))
	canvas := letterbox(img, d.size)
	input := toCHW(canvas, false)
	for i := range input {
		input[i] = (input[i] - 127.5) / 128
	}
	outs, err := d.model.run(input)
	if err != nil {
		return nil, err
	}
	boxes, err := decodeSCRFD(outs, scale, threshold)
	if err != nil {
		return nil, err
	}
	return nms(boxes, nmsThreshold), nil
}

// scrfdStride 返回锚点数量对应的下采样步长。
func scrfdStride(anchors int) (int, bool) {
	switch anchors {
	case 80 * 80 * 2:
		return 8, true
	case 40 * 40 * 2:
		return 16, true
	case 20 * 20 * 2:
		return 32, true
	default:
		return 0, false
	}
}

// decodeSCRFD 按 2 anchor/位置解码分数、边界距离与关键点。
func decodeSCRFD(outs []*tensors.Tensor, scale, threshold float64) ([]domain.FaceBox, error) {
	type heads struct {
		score, bbox, kps []float32
	}
	byStride := map[int]*heads{}
	for _, t := range outs {
		dims := t.Shape().Dimensions
		if len(dims) != 2 || dims[1] == 0 {
			return nil, fmt.Errorf("SCRFD 输出维度异常: %v", dims)
		}
		stride, ok := scrfdStride(dims[0])
		if !ok {
			return nil, fmt.Errorf("SCRFD 锚点数量未知: %d", dims[0])
		}
		h := byStride[stride]
		if h == nil {
			h = &heads{}
			byStride[stride] = h
		}
		flat := flatF32(t)
		switch dims[1] {
		case 1:
			h.score = flat
		case 4:
			h.bbox = flat
		case 10:
			h.kps = flat
		default:
			return nil, fmt.Errorf("SCRFD 输出宽度未知: %d", dims[1])
		}
	}

	var boxes []domain.FaceBox
	for _, stride := range []int{8, 16, 32} {
		h := byStride[stride]
		if h == nil || h.score == nil || h.bbox == nil || h.kps == nil {
			return nil, fmt.Errorf("SCRFD stride %d 输出缺失", stride)
		}
		cells := int(math.Sqrt(float64(len(h.score) / 2)))
		anchors := cells * cells * 2
		if len(h.score) != anchors {
			return nil, fmt.Errorf("SCRFD stride %d 锚点数不一致: %d", stride, len(h.score))
		}
		for i := 0; i < anchors; i++ {
			score := float64(h.score[i])
			if math.IsNaN(score) || score < threshold {
				continue
			}
			cellsIdx := i / 2
			cx := float64((cellsIdx%cells)*stride) / scale
			cy := float64((cellsIdx/cells)*stride) / scale
			d := h.bbox[i*4:]
			box := domain.FaceBox{
				X1:           cx - float64(d[0])*float64(stride)/scale,
				Y1:           cy - float64(d[1])*float64(stride)/scale,
				X2:           cx + float64(d[2])*float64(stride)/scale,
				Y2:           cy + float64(d[3])*float64(stride)/scale,
				Score:        score,
				HasLandmarks: true,
			}
			k := h.kps[i*10:]
			for j := 0; j < 5; j++ {
				box.Landmarks[j] = [2]float64{
					cx + float64(k[j*2])*float64(stride)/scale,
					cy + float64(k[j*2+1])*float64(stride)/scale,
				}
			}
			boxes = append(boxes, box)
		}
	}
	return boxes, nil
}
