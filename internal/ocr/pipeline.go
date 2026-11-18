package ocr

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"

	"golang.org/x/image/draw"
)

// Box 是原图坐标系下的文本框。
type Box struct {
	X, Y, W, H int
}

// Line 是同一行的识别结果。
type Line struct {
	Text       string
	Confidence float64
	Box        Box
}

// Result 是一次 OCR 的完整结果。
type Result struct {
	Text        string
	Lines       [][]Line
	Confidence  float64
	Boxes       []Box
	Orientation int
	DetDuration int64 // 纳秒
	RecDuration int64
	Total       int64
}

type candidate struct {
	text       string
	confidence float64
	box        Box
}

const detNonZeroThreshold = 0.5 / 255.0

// detect 对齐 ppu-paddle-ocr：概率图 round(p*255) -> 轮廓 boundingRect -> 按高度外扩 -> 映射回原图。
func (e *Engine) detect(src image.Image) ([]Box, float64, error) {
	sw, sh := src.Bounds().Dx(), src.Bounds().Dy()
	if sw <= 0 || sh <= 0 {
		return nil, 1, fmt.Errorf("空图像")
	}
	ratio := 1.0
	if m := math.Max(float64(sw), float64(sh)); m > float64(e.cfg.MaxSide) {
		ratio = float64(e.cfg.MaxSide) / m
	}
	rw := max(1, int(math.Round(float64(sw)*ratio)))
	rh := max(1, int(math.Round(float64(sh)*ratio)))
	padW := int(math.Ceil(float64(rw)/32)) * 32
	padH := int(math.Ceil(float64(rh)/32)) * 32

	resized := image.NewRGBA(image.Rect(0, 0, padW, padH))
	draw.BiLinear.Scale(resized, image.Rect(0, 0, rw, rh), src, src.Bounds(), draw.Over, nil)

	mean := [3]float32{0.485, 0.456, 0.406}
	std := [3]float32{0.229, 0.224, 0.225}
	data := make([]float32, 3*padW*padH)
	for y := 0; y < padH; y++ {
		for x := 0; x < padW; x++ {
			r, g, b, _ := resized.At(x, y).RGBA()
			px := [3]float32{float32(r) / 65535, float32(g) / 65535, float32(b) / 65535}
			for c := 0; c < 3; c++ {
				data[c*padW*padH+y*padW+x] = (px[c] - mean[c]) / std[c]
			}
		}
	}

	out, dims, err := e.det.run(data, 1, 3, padH, padW)
	if err != nil {
		return nil, ratio, err
	}
	if len(dims) != 4 {
		return nil, ratio, fmt.Errorf("检测输出维度异常: %v", dims)
	}
	oh, ow := dims[2], dims[3]
	if oh <= 0 || ow <= 0 || len(out) < oh*ow {
		return nil, ratio, fmt.Errorf("检测输出尺寸异常: %v", dims)
	}

	mask := make([]bool, ow*oh)
	for i := 0; i < ow*oh; i++ {
		mask[i] = float64(out[i]) >= detNonZeroThreshold
	}

	rects := contourBoxes(mask, ow, oh)
	boxes := make([]Box, 0, len(rects))
	for _, rect := range rects {
		if rect.W*rect.H <= e.cfg.MinBoxArea {
			continue
		}
		padded := padRect(rect, ow, oh, e.cfg.PaddingVertical, e.cfg.PaddingHorizontal)
		final := toOriginal(padded, ratio, sw, sh)
		if final.W > 5 && final.H > 5 {
			boxes = append(boxes, final)
		}
	}
	return boxes, ratio, nil
}

// contourBoxes 近似 OpenCV findContours(RETR_LIST)：前景 8 连通 + 不触边背景孔洞。
func contourBoxes(mask []bool, w, h int) []Box {
	boxes := connected(mask, w, h, true, false)
	holes := connected(mask, w, h, false, true)
	return append(boxes, holes...)
}

// connected 在二值掩码上提取 8 连通区域；skipBorder 时丢弃触边区域（用于孔洞）。
func connected(mask []bool, w, h int, foreground, skipBorder bool) []Box {
	seen := make([]bool, len(mask))
	boxes := make([]Box, 0)
	stack := make([]int, 0, 256)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			idx := y*w + x
			if seen[idx] || mask[idx] != foreground {
				continue
			}
			minX, minY, maxX, maxY := x, y, x, y
			touchBorder := false
			stack = stack[:0]
			stack = append(stack, idx)
			seen[idx] = true
			for len(stack) > 0 {
				cur := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				cy, cx := cur/w, cur%w
				if cx < minX {
					minX = cx
				}
				if cx > maxX {
					maxX = cx
				}
				if cy < minY {
					minY = cy
				}
				if cy > maxY {
					maxY = cy
				}
				if cx == 0 || cy == 0 || cx == w-1 || cy == h-1 {
					touchBorder = true
				}
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						nx, ny := cx+dx, cy+dy
						if nx < 0 || ny < 0 || nx >= w || ny >= h {
							continue
						}
						ni := ny*w + nx
						if !seen[ni] && mask[ni] == foreground {
							seen[ni] = true
							stack = append(stack, ni)
						}
					}
				}
			}
			if skipBorder && touchBorder {
				continue
			}
			boxes = append(boxes, Box{X: minX, Y: minY, W: maxX - minX + 1, H: maxY - minY + 1})
		}
	}
	return boxes
}

// padRect 按参考实现以框高为基准外扩，并裁剪到画布边界。
func padRect(rect Box, maxW, maxH int, padV, padH float64) Box {
	vPad := int(math.Round(float64(rect.H) * padV))
	hPad := int(math.Round(float64(rect.H) * padH))
	x := max(0, rect.X-hPad)
	y := max(0, rect.Y-vPad)
	right := min(maxW, rect.X+rect.W+hPad)
	bottom := min(maxH, rect.Y+rect.H+vPad)
	return Box{X: x, Y: y, W: right - x, H: bottom - y}
}

// toOriginal 把缩放坐标还原到原图，并限制在图像边界内。
func toOriginal(rect Box, ratio float64, origW, origH int) Box {
	x := max(0, int(math.Round(float64(rect.X)/ratio)))
	y := max(0, int(math.Round(float64(rect.Y)/ratio)))
	w := min(origW-x, int(math.Round(float64(rect.W)/ratio)))
	h := min(origH-y, int(math.Round(float64(rect.H)/ratio)))
	return Box{X: x, Y: y, W: w, H: h}
}

// recognize 对齐 BaseRecognitionService：等比高 48、R 通道、CTC 贪心。
func (e *Engine) recognize(src image.Image, b Box) (string, float64, error) {
	if b.W <= 0 || b.H <= 0 {
		return "", 0, nil
	}
	tw := int(math.Round(float64(b.W) * float64(recHeight) / float64(b.H)))
	if tw < recMinWidth {
		return "", 0, nil
	}
	targetW := tw
	if e.cfg.WidthBucket > 0 {
		targetW = int(math.Ceil(float64(tw)/float64(e.cfg.WidthBucket))) * e.cfg.WidthBucket
	}
	if targetW < recMinWidth {
		targetW = recMinWidth
	}
	crop := image.NewRGBA(image.Rect(0, 0, targetW, recHeight))
	draw.Draw(crop, crop.Bounds(), image.NewUniform(color.RGBA{127, 127, 127, 255}), image.Point{}, draw.Src)
	draw.BiLinear.Scale(crop, image.Rect(0, 0, tw, recHeight), src, image.Rect(b.X, b.Y, b.X+b.W, b.Y+b.H), draw.Over, nil)

	data := make([]float32, 3*recHeight*targetW)
	for y := 0; y < recHeight; y++ {
		for x := 0; x < targetW; x++ {
			r, _, _, _ := crop.At(x, y).RGBA()
			v := float32(r)/65535*2 - 1
			for c := 0; c < 3; c++ {
				data[c*recHeight*targetW+y*targetW+x] = v
			}
		}
	}

	out, dims, err := e.rec.run(data, 1, 3, recHeight, targetW)
	if err != nil {
		return "", 0, err
	}
	if len(dims) != 3 {
		return "", 0, fmt.Errorf("识别输出维度异常: %v", dims)
	}
	seq, classes := dims[1], dims[2]
	if len(out) < seq*classes {
		return "", 0, fmt.Errorf("识别输出长度不足")
	}
	var sb strings.Builder
	last := 0
	sum, n := 0.0, 0
	for t := 0; t < seq; t++ {
		base := t * classes
		best, bestI := -math.MaxFloat32, 0
		for c := 0; c < classes; c++ {
			if v := float64(out[base+c]); v > best {
				best, bestI = v, c
			}
		}
		if bestI == 0 || bestI == last {
			last = bestI
			continue
		}
		if bestI == len(e.dict)-1 {
			sb.WriteRune(' ')
		} else if bestI >= 0 && bestI < len(e.dict) {
			sb.WriteRune(e.dict[bestI])
		}
		sum += best
		n++
		last = bestI
	}
	conf := 0.0
	if n > 0 {
		conf = sum / float64(n)
	}
	return sb.String(), conf, nil
}

// sortReadingOrder 按阅读顺序排序：同一行按 x，否则按 y。
func sortReadingOrder(results []candidate) {
	sort.SliceStable(results, func(i, j int) bool {
		a, b := results[i].box, results[j].box
		if max(a.Y-b.Y, b.Y-a.Y) < max(1, (a.H+b.H)/4) {
			return a.X < b.X
		}
		return a.Y < b.Y
	})
}

// groupResultsByLine 按 y 邻近度分行，并在行内按 x 排序。
func groupResultsByLine(results []candidate) [][]Line {
	if len(results) == 0 {
		return nil
	}
	lines := make([][]Line, 0)
	current := []candidate{results[0]}
	currentY := results[0].box.Y
	avgHeight := float64(results[0].box.H)
	flush := func() {
		sort.SliceStable(current, func(i, j int) bool { return current[i].box.X < current[j].box.X })
		line := make([]Line, 0, len(current))
		for _, r := range current {
			line = append(line, Line{Text: r.text, Confidence: r.confidence, Box: r.box})
		}
		lines = append(lines, line)
	}
	for _, r := range results[1:] {
		if math.Abs(float64(r.box.Y-currentY)) < avgHeight/2 {
			current = append(current, r)
			avgHeight = (avgHeight*float64(len(current)-1) + float64(r.box.H)) / float64(len(current))
		} else {
			flush()
			current = []candidate{r}
			currentY = r.box.Y
			avgHeight = float64(r.box.H)
		}
	}
	flush()
	return lines
}

// joinLines 拼接行文本并计算字符级平均置信度。
func joinLines(lines [][]Line) (string, float64) {
	parts := make([]string, 0, len(lines))
	total, n := 0.0, 0
	for _, line := range lines {
		texts := make([]string, 0, len(line))
		for _, r := range line {
			texts = append(texts, r.Text)
			total += r.Confidence
			n++
		}
		parts = append(parts, strings.Join(texts, " "))
	}
	conf := 0.0
	if n > 0 {
		conf = total / float64(n)
	}
	return strings.Join(parts, "\n"), conf
}

// rotate 按 0/90/180/270 顺时针旋转图像。
func rotate(src image.Image, angle int) image.Image {
	if angle == 0 {
		return src
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	var dst *image.RGBA
	switch angle {
	case 90:
		dst = image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(h-1-y, x, src.At(x, y))
			}
		}
	case 180:
		dst = image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(w-1-x, h-1-y, src.At(x, y))
			}
		}
	case 270:
		dst = image.NewRGBA(image.Rect(0, 0, h, w))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				dst.Set(y, w-1-x, src.At(x, y))
			}
		}
	default:
		return src
	}
	return dst
}
